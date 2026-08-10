package prfix

type pageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type reviewComment struct {
	ID     string `json:"id"`
	Body   string `json:"body"`
	URL    string `json:"url"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

type commentConnection struct {
	Nodes    []reviewComment `json:"nodes"`
	PageInfo pageInfo        `json:"pageInfo"`
}

type reviewThread struct {
	ID         string            `json:"id"`
	IsResolved bool              `json:"isResolved"`
	IsOutdated bool              `json:"isOutdated"`
	Path       string            `json:"path"`
	Line       int               `json:"line"`
	StartLine  int               `json:"startLine"`
	Comments   commentConnection `json:"comments"`
}

type pullRequestReview struct {
	ID     string `json:"id"`
	State  string `json:"state"`
	Body   string `json:"body"`
	URL    string `json:"url"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
}

type issueComment struct {
	ID                string `json:"id"`
	Body              string `json:"body"`
	URL               string `json:"url"`
	AuthorAssociation string `json:"authorAssociation"`
	Author            struct {
		Login string `json:"login"`
	} `json:"author"`
}

type threadPageResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes    []reviewThread `json:"nodes"`
					PageInfo pageInfo       `json:"pageInfo"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

type threadCommentPageResponse struct {
	Data struct {
		Node struct {
			Comments commentConnection `json:"comments"`
		} `json:"node"`
	} `json:"data"`
}

type reviewPageResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				Reviews struct {
					Nodes    []pullRequestReview `json:"nodes"`
					PageInfo pageInfo            `json:"pageInfo"`
				} `json:"reviews"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

type issueCommentPageResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes    []issueComment `json:"nodes"`
					PageInfo pageInfo       `json:"pageInfo"`
				} `json:"comments"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}
