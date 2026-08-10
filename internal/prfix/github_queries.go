package prfix

import (
	"fmt"
	"sort"
)

func threadQuery(pageSize int) string {
	return fmt.Sprintf(`query($owner:String!,$name:String!,$number:Int!,$cursor:String){
  repository(owner:$owner,name:$name){pullRequest(number:$number){
    reviewThreads(first:%d,after:$cursor){pageInfo{hasNextPage endCursor} nodes{
      id isResolved isOutdated path line startLine
      comments(first:%d){pageInfo{hasNextPage endCursor} nodes{id body url author{login}}}
    }}
  }}
}`, pageSize, pageSize)
}

func threadCommentQuery(pageSize int) string {
	return fmt.Sprintf(`query($threadId:ID!,$cursor:String){
  node(id:$threadId){... on PullRequestReviewThread{
    comments(first:%d,after:$cursor){pageInfo{hasNextPage endCursor} nodes{id body url author{login}}}
  }}
}`, pageSize)
}

func reviewQuery(pageSize int) string {
	return fmt.Sprintf(`query($owner:String!,$name:String!,$number:Int!,$cursor:String){
  repository(owner:$owner,name:$name){pullRequest(number:$number){
    reviews(first:%d,after:$cursor,states:[CHANGES_REQUESTED]){
      pageInfo{hasNextPage endCursor} nodes{id state body url author{login}}
    }
  }}
}`, pageSize)
}

func issueCommentQuery(pageSize int) string {
	return fmt.Sprintf(`query($owner:String!,$name:String!,$number:Int!,$cursor:String){
  repository(owner:$owner,name:$name){pullRequest(number:$number){
    comments(first:%d,after:$cursor){pageInfo{hasNextPage endCursor} nodes{
      id body url authorAssociation author{login}
    }}
  }}
}`, pageSize)
}

func graphQLArgs(query string, variables map[string]string) []string {
	args := []string{"api", "graphql", "-f", "query=" + query}
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		flag := "-f"
		if key == "number" {
			flag = "-F"
		}
		args = append(args, flag, key+"="+variables[key])
	}
	return args
}
