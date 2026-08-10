package usage

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var shardPattern = regexp.MustCompile(`^(\d{4}-\d{2})-(\d{4})\.jsonl$`)

type shard struct {
	path string
	name string
	size int64
}

func Record(input RecordInput) error {
	settings, err := EffectiveSettings(input.ProjectRoot)
	if err != nil || !settings.Enabled || strings.HasPrefix(input.Command, "usage") {
		return err
	}
	dir, err := Directory()
	if err != nil {
		return err
	}
	return withStoreLock(dir, func() error {
		if err := ensurePrivateDirectory(dir); err != nil {
			return err
		}
		projectID, err := projectIdentifier(dir, input.ProjectRoot, true)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		event := Event{
			SchemaVersion: SchemaVersion,
			Timestamp:     now,
			Command:       strings.TrimSpace(input.Command),
			Version:       strings.TrimSpace(input.Version),
			ExitCode:      input.ExitCode,
			Success:       input.ExitCode == 0,
			ElapsedMS:     max(input.Elapsed.Milliseconds(), 0),
			ProjectID:     projectID,
			Interactive:   input.Interactive,
		}
		line, err := json.Marshal(event)
		if err != nil {
			return err
		}
		line = append(line, '\n')
		if err := automaticMaintenance(dir, now); err != nil {
			return err
		}
		path, err := writableShard(dir, now, int64(len(line)))
		if err != nil {
			return err
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		if _, err := file.Write(line); err != nil {
			_ = file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
		return pruneTotalBytes(dir, false, nil)
	})
}

func ensurePrivateDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.Chmod(dir, 0o700)
}

func writableShard(dir string, now time.Time, incoming int64) (string, error) {
	prefix := now.UTC().Format("2006-01")
	shards, err := listShards(dir)
	if err != nil {
		return "", err
	}
	index := 1
	for _, item := range shards {
		match := shardPattern.FindStringSubmatch(item.name)
		if len(match) != 3 || match[1] != prefix {
			continue
		}
		value, _ := strconv.Atoi(match[2])
		if value > index {
			index = value
		}
		if value == index && item.size+incoming <= MaxShardBytes {
			return item.path, nil
		}
	}
	if len(shards) > 0 {
		last := shards[len(shards)-1]
		match := shardPattern.FindStringSubmatch(last.name)
		if len(match) == 3 && match[1] == prefix {
			value, _ := strconv.Atoi(match[2])
			if last.size+incoming <= MaxShardBytes {
				return last.path, nil
			}
			index = value + 1
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%04d.jsonl", prefix, index)), nil
}

func listShards(dir string) ([]shard, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []shard{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := make([]shard, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !shardPattern.MatchString(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		items = append(items, shard{path: filepath.Join(dir, entry.Name()), name: entry.Name(), size: info.Size()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	return items, nil
}

func readEvents(path string, visit func(Event) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	reader := bufio.NewReader(file)
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(strings.TrimSpace(string(line))) > 0 {
			var event Event
			if err := json.Unmarshal(line, &event); err != nil {
				return fmt.Errorf("invalid usage event in %s: %w", filepath.Base(path), err)
			}
			if event.SchemaVersion != SchemaVersion {
				return fmt.Errorf("unsupported usage event schema %q in %s", event.SchemaVersion, filepath.Base(path))
			}
			if err := visit(event); err != nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func projectIdentifier(dir, root string, create bool) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", nil
	}
	identityPath := filepath.Join(dir, ".identity")
	data, err := os.ReadFile(identityPath)
	if errors.Is(err, os.ErrNotExist) && create {
		data = make([]byte, 32)
		if _, err := rand.Read(data); err != nil {
			return "", err
		}
		if err := os.WriteFile(identityPath, []byte(hex.EncodeToString(data)), 0o600); err != nil {
			return "", err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	if decoded, decodeErr := hex.DecodeString(strings.TrimSpace(string(data))); decodeErr == nil {
		data = decoded
	}
	digest := sha256.Sum256(append(append([]byte{}, data...), []byte("\x00"+filepath.Clean(root))...))
	return hex.EncodeToString(digest[:16]), nil
}
