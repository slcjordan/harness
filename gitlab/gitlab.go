package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func deref[T any](t *T) (result T) {
	if t != nil {
		result = *t
	}
	return
}

func userIDs(in []*gitlab.BasicUser) []int32 {
	var result []int32
	for _, a := range in {
		result = append(result, int32(deref(a).ID))
	}
	return result
}

var parseJiraRegexp = regexp.MustCompile("^(Resolve )?([A-Z]+-[0-9]+) .*$")

func parseJira(title string) string {
	groups := parseJiraRegexp.FindStringSubmatch(title)
	if len(groups) < 3 {
		return ""
	}
	return groups[2]
}

func mergeRequest(in gitlab.BasicMergeRequest) harness.MergeRequest {
	return harness.MergeRequest{
		ID:                          in.ID,
		IID:                         in.IID,
		TargetBranch:                in.TargetBranch,
		SourceBranch:                in.SourceBranch,
		ProjectID:                   in.ProjectID,
		Jira:                        parseJira(in.Title),
		Title:                       in.Title,
		State:                       in.State,
		Imported:                    in.Imported,
		ImportedFrom:                in.ImportedFrom,
		CreatedAt:                   deref(in.CreatedAt),
		UpdatedAt:                   deref(in.UpdatedAt),
		Upvotes:                     in.Upvotes,
		Downvotes:                   in.Downvotes,
		Author:                      int32(deref(in.Author).ID),
		Assignee:                    deref(in.Assignee).ID,
		Assignees:                   userIDs(in.Assignees),
		Reviewers:                   userIDs(in.Reviewers),
		SourceProjectID:             in.SourceProjectID,
		TargetProjectID:             in.TargetProjectID,
		Labels:                      in.Labels,
		Description:                 in.Description,
		Draft:                       in.Draft,
		MergeWhenPipelineSucceeds:   in.MergeWhenPipelineSucceeds,
		DetailedMergeStatus:         in.DetailedMergeStatus,
		MergeUser:                   int32(deref(in.MergeUser).ID),
		MergedAt:                    deref(in.MergedAt),
		MergeAfter:                  deref(in.MergeAfter),
		PreparedAt:                  deref(in.PreparedAt),
		ClosedBy:                    int32(deref(in.ClosedBy).ID),
		ClosedAt:                    deref(in.ClosedAt),
		SHA:                         in.SHA,
		MergeCommitSHA:              in.MergeCommitSHA,
		SquashCommitSHA:             in.SquashCommitSHA,
		UserNotesCount:              in.UserNotesCount,
		ShouldRemoveSourceBranch:    in.ShouldRemoveSourceBranch,
		ForceRemoveSourceBranch:     in.ForceRemoveSourceBranch,
		AllowCollaboration:          in.AllowCollaboration,
		AllowMaintainerToPush:       in.AllowMaintainerToPush,
		WebURL:                      in.WebURL,
		ShortReference:              in.References.Short,
		RelativeReference:           in.References.Relative,
		FullReference:               in.References.Full,
		DiscussionLocked:            in.DiscussionLocked,
		Squash:                      in.Squash,
		SquashOnMerge:               in.SquashOnMerge,
		TaskCount:                   in.TaskCompletionStatus.Count,
		TaskCompletedCount:          in.TaskCompletionStatus.CompletedCount,
		HasConflicts:                in.HasConflicts,
		BlockingDiscussionsResolved: in.BlockingDiscussionsResolved,
	}
}

func mergeRequests(mrs []*gitlab.BasicMergeRequest) []harness.MergeRequest {
	var result []harness.MergeRequest
	for _, mr := range mrs {
		result = append(result, mergeRequest(deref(mr)))
	}
	return result
}

type ListMRs struct {
	Client *gitlab.Client
}

func (l *ListMRs) Handle(ctx context.Context, input string) ([]harness.MergeRequest, error) {
	user, _, err := l.Client.Users.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("could not get current authenticated user: %s", err)
	}
	mrList, _, err := l.Client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		AuthorID: gitlab.Ptr(deref(user).ID),
		// Optional filters
		State: gitlab.Ptr("opened"), // "opened", "closed", "merged", or "all"
	})
	if err != nil {
		return nil, fmt.Errorf("could not list current users' MRs: %s", err)
	}
	/*
		mrs := make([]*gitlab.MergeRequest, 0, len(mrList))
		for _, mr := range mrList {
			curr, _, err := l.Client.MergeRequests.GetMergeRequest(mr.ProjectID, mr.IID, nil)
			if err != nil {
				return nil, fmt.Errorf("could not get mr %s-%d details: %s", mr.ProjectID, mr.IID, err)
			}
			mrs = append(mrs, curr)
		}
	*/

	return mergeRequests(mrList), nil
}

type UserMessages struct {
	Client     *gitlab.Client
	CacheFetch harness.Handler[int32, harness.GitlabUser]
	CacheSave  harness.Handler[harness.GitlabUser, struct{}]
}

func (u *UserMessages) problems(ctx context.Context, curr harness.MergeRequest) []string {
	var result []string
	if !curr.BlockingDiscussionsResolved || curr.HasConflicts {
		result = append(result, "has unresolved comments/conflicts")
	}
	if len(curr.Reviewers) < 2 {
		result = append(result, "needs 2 or more reviewers")
	}
	return result
}

func (u *UserMessages) Handle(ctx context.Context, mrs []harness.MergeRequest) ([]harness.UserMessage, error) {
	msgs := make(map[int32]*bytes.Buffer)
	unresolved := make(map[string]bool)

	// sort.SliceStable(mrs, func(i, j int) bool { return mrs[i].Assigned < mrs[j].Assigned })
	sort.SliceStable(mrs, func(i, j int) bool { return mrs[i].Jira < mrs[j].Jira })

	for _, curr := range mrs {
		w, ok := msgs[curr.Author]
		if !ok {
			w = bytes.NewBuffer(nil)
			msgs[curr.Author] = w
		}
		problems := u.problems(ctx, curr)
		if len(problems) > 0 {
			if !unresolved[curr.Jira] {
				unresolved[curr.Jira] = true
				fmt.Fprintf(w, "\n*%s*\n\n", curr.Jira)
			}
			name, _ := u.projectName(curr.ProjectID)
			fmt.Fprintf(w, "<%s|%s> %s.\n", curr.WebURL, name, strings.Join(problems, " and "))
		}
	}

	var result []harness.UserMessage
	for userID, msg := range msgs {
		email, err := u.getEmail(ctx, userID)
		if err != nil {
			logger.Errorf(ctx, "could not get email: %s", err)
			continue
		}
		result = append(result, harness.UserMessage{
			Email:   email,
			Message: msg.String(),
		})
	}
	return result, nil
}

func (u *UserMessages) getEmail(ctx context.Context, userID int32) (string, error) {
	user, err := u.CacheFetch.Handle(ctx, userID)
	found := (err == nil)
	if found {
		return user.Email, nil
	}
	userMRList, _, err := u.Client.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		AuthorID: gitlab.Ptr(int(userID)),
	})
	if err != nil {
		return "", err
	}

	var commits []*gitlab.Commit
	for len(commits) == 0 {
		if len(userMRList) == 0 {
			return "", fmt.Errorf("could not find any mrs for %d", userID)
		}
		first := userMRList[0]
		commits, _, err = u.Client.MergeRequests.GetMergeRequestCommits(first.ProjectID, first.IID, nil)
		if err != nil {
			return "", err
		}
		userMRList = userMRList[1:]
	}

	user = harness.GitlabUser{
		ID:    userID,
		Email: deref(commits[0]).AuthorEmail,
	}
	_, err = u.CacheSave.Handle(ctx, user)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}

func (u *UserMessages) projectName(projectID int) (string, string) {
	project, _, err := u.Client.Projects.GetProject(projectID, nil)
	if err != nil {
		return "", ""
	}
	return project.Name, project.PathWithNamespace
}
