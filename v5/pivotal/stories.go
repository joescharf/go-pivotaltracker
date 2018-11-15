// Copyright (c) 2014 Salsita Software
// Copyright (C) 2015 Scott Devoid
// Use of this source code is governed by the MIT License.
// The license can be found in the LICENSE file.

package pivotal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Number of items to fetch at once when getting paginated response.
const PageLimit = 10

const (
	StoryTypeFeature = "feature"
	StoryTypeBug     = "bug"
	StoryTypeChore   = "chore"
	StoryTypeRelease = "release"
)

const (
	StoryStateUnscheduled = "unscheduled"
	StoryStatePlanned     = "planned"
	StoryStateUnstarted   = "unstarted"
	StoryStateStarted     = "started"
	StoryStateFinished    = "finished"
	StoryStateDelivered   = "delivered"
	StoryStateAccepted    = "accepted"
	StoryStateRejected    = "rejected"
)

type Story struct {
	Id            int        `json:"id,omitempty"`
	ProjectId     int        `json:"project_id,omitempty"`
	Name          string     `json:"name,omitempty"`
	Description   string     `json:"description,omitempty"`
	Type          string     `json:"story_type,omitempty"`
	State         string     `json:"current_state,omitempty"`
	Estimate      *float64   `json:"estimate,omitempty"`
	AcceptedAt    *time.Time `json:"accepted_at,omitempty"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	RequestedById int        `json:"requested_by_id,omitempty"`
	OwnerIds      []int      `json:"owner_ids,omitempty"`
	LabelIds      []int      `json:"label_ids,omitempty"`
	Labels        []*Label   `json:"labels,omitempty"`
	TaskIds       []int      `json:"task_ids,omitempty"`
	Tasks         []int      `json:"tasks,omitempty"`
	FollowerIds   []int      `json:"follower_ids,omitempty"`
	CommentIds    []int      `json:"comment_ids,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	IntegrationId int        `json:"integration_id,omitempty"`
	ExternalId    string     `json:"external_id,omitempty"`
	URL           string     `json:"url,omitempty"`
}

type StoryRequest struct {
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"story_type,omitempty"`
	State       string    `json:"current_state,omitempty"`
	Estimate    *float64  `json:"estimate,omitempty"`
	OwnerIds    *[]int    `json:"owner_ids,omitempty"`
	LabelIds    *[]int    `json:"label_ids,omitempty"`
	Labels      *[]*Label `json:"labels,omitempty"`
	TaskIds     *[]int    `json:"task_ids,omitempty"`
	Tasks       *[]int    `json:"tasks,omitempty"`
	FollowerIds *[]int    `json:"follower_ids,omitempty"`
	CommentIds  *[]int    `json:"comment_ids,omitempty"`
}

type Label struct {
	Id        int        `json:"id,omitempty"`
	ProjectId int        `json:"project_id,omitempty"`
	Name      string     `json:"name,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Kind      string     `json:"kind,omitempty"`
}

type Task struct {
	Id          int        `json:"id,omitempty"`
	StoryId     int        `json:"story_id,omitempty"`
	Description string     `json:"description,omitempty"`
	Position    int        `json:"position,omitempty"`
	Complete    bool       `json:"complete,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type Person struct {
	Id       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Initials string `json:"initials,omitempty"`
	Username string `json:"username,omitempty"`
	Kind     string `json:"kind,omitempty"`
}

type Comment struct {
	Id                  int        `json:"id,omitempty"`
	StoryId             int        `json:"story_id,omitempty"`
	EpicId              int        `json:"epic_id,omitempty"`
	PersonId            int        `json:"person_id,omitempty"`
	Text                string     `json:"text,omitempty"`
	FileAttachmentIds   []int      `json:"file_attachment_ids,omitempty"`
	GoogleAttachmentIds []int      `json:"google_attachment_ids,omitempty"`
	CommitType          string     `json:"commit_type,omitempty"`
	CommitIdentifier    string     `json:"commit_identifier,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type Blocker struct {
	Id          int        `json:"id,omitempty"`
	StoryId     int        `json:"story_id,omitempty"`
	PersonId    int        `json:"person_id,omitempty"`
	Description string     `json:"description,omitempty"`
	Resolved    bool       `json:"resolved,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type BlockerRequest struct {
	Description string `json:"description,omitempty"`
	Resolved    *bool  `json:"resolved,omitempty"`
}

type StoryService struct {
	client *Client
}

func newStoryService(client *Client) *StoryService {
	return &StoryService{client}
}

// List returns all stories matching the filter in case the filter is specified.
//
// List actually sends 2 HTTP requests - one to get the total number of stories,
// another to retrieve the stories using the right pagination setup. The reason
// for this is that the filter might require to fetch all the stories at once
// to get the right results. Since the response as generated by Pivotal Tracker
// is not always sorted when using a filter, this approach is required to get
// the right data. Not sure whether this is a bug or a feature.
func (service *StoryService) List(projectId int, filter string) ([]*Story, error) {
	reqFunc := newStoriesRequestFunc(service.client, projectId, filter)
	cursor, err := newCursor(service.client, reqFunc, 0)
	if err != nil {
		return nil, err
	}

	var stories []*Story
	if err := cursor.all(&stories); err != nil {
		return nil, err
	}
	return stories, nil
}

func newStoriesRequestFunc(client *Client, projectId int, filter string) func() *http.Request {
	return func() *http.Request {
		u := fmt.Sprintf("projects/%v/stories", projectId)
		if filter != "" {
			u += "?filter=" + url.QueryEscape(filter)
		}
		req, _ := client.NewRequest("GET", u, nil)
		return req
	}
}

type StoryCursor struct {
	*cursor
	buff []*Story
}

// Next returns the next story.
//
// In case there are no more stories, io.EOF is returned as an error.
func (c *StoryCursor) Next() (s *Story, err error) {
	if len(c.buff) == 0 {
		_, err = c.next(&c.buff)
		if err != nil {
			return nil, err
		}
	}

	if len(c.buff) == 0 {
		err = io.EOF
	} else {
		s, c.buff = c.buff[0], c.buff[1:]
	}
	return s, err
}

// Iterate returns a cursor that can be used to iterate over the stories specified
// by the filter. More stories are fetched on demand as needed.
func (service *StoryService) Iterate(projectId int, filter string) (c *StoryCursor, err error) {
	reqFunc := newStoriesRequestFunc(service.client, projectId, filter)
	cursor, err := newCursor(service.client, reqFunc, PageLimit)
	if err != nil {
		return nil, err
	}
	return &StoryCursor{cursor, make([]*Story, 0)}, nil
}

func (service *StoryService) Create(projectId int, story *StoryRequest) (*Story, *http.Response, error) {
	if projectId == 0 {
		return nil, nil, &ErrFieldNotSet{"project_id"}
	}

	if story.Name == "" {
		return nil, nil, &ErrFieldNotSet{"name"}
	}

	u := fmt.Sprintf("projects/%v/stories", projectId)
	req, err := service.client.NewRequest("POST", u, story)
	if err != nil {
		return nil, nil, err
	}

	var newStory Story

	resp, err := service.client.Do(req, &newStory)
	if err != nil {
		return nil, resp, err
	}

	return &newStory, resp, nil
}

func (service *StoryService) Get(projectId, storyId int) (*Story, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v", projectId, storyId)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var story Story
	resp, err := service.client.Do(req, &story)
	if err != nil {
		return nil, resp, err
	}

	return &story, resp, err
}

func arrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func (service *StoryService) GetBulk(projectId int, storyIds []int) ([]*Story, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/bulk", projectId)
	stories := arrayToString(storyIds, ",")

	if stories != "" {
		u += "?ids=" + url.QueryEscape(stories)
	}

	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var bulkStories []*Story
	resp, err := service.client.Do(req, &bulkStories)
	if err != nil {
		return nil, resp, err
	}

	return bulkStories, resp, err
}

func (service *StoryService) Update(projectId, storyId int, story *StoryRequest) (*Story, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v", projectId, storyId)
	req, err := service.client.NewRequest("PUT", u, story)
	if err != nil {
		return nil, nil, err
	}

	var bodyStory Story
	resp, err := service.client.Do(req, &bodyStory)
	if err != nil {
		return nil, resp, err
	}

	return &bodyStory, resp, err

}

func (service *StoryService) ListTasks(projectId, storyId int) ([]*Task, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v/tasks", projectId, storyId)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*Task
	resp, err := service.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, err
}

func (service *StoryService) AddTask(projectId, storyId int, task *Task) (*http.Response, error) {
	if task.Description == "" {
		return nil, &ErrFieldNotSet{"description"}
	}

	u := fmt.Sprintf("projects/%v/stories/%v/tasks", projectId, storyId)
	req, err := service.client.NewRequest("POST", u, task)
	if err != nil {
		return nil, err
	}

	return service.client.Do(req, nil)
}

func (service *StoryService) ListOwners(projectId, storyId int) ([]*Person, *http.Response, error) {
	u := fmt.Sprintf("projects/%d/stories/%d/owners", projectId, storyId)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var owners []*Person
	resp, err := service.client.Do(req, &owners)
	if err != nil {
		return nil, resp, err
	}

	return owners, resp, err
}

func (service *StoryService) AddComment(
	projectId int,
	storyId int,
	comment *Comment,
) (*Comment, *http.Response, error) {

	u := fmt.Sprintf("projects/%v/stories/%v/comments", projectId, storyId)
	req, err := service.client.NewRequest("POST", u, comment)
	if err != nil {
		return nil, nil, err
	}

	var newComment Comment
	resp, err := service.client.Do(req, &newComment)
	if err != nil {
		return nil, resp, err
	}

	return &newComment, resp, err
}

// ListComments returns the list of Comments in a Story.
func (service *StoryService) ListComments(
	projectId int,
	storyId int,
) ([]*Comment, *http.Response, error) {

	u := fmt.Sprintf("projects/%v/stories/%v/comments", projectId, storyId)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var comments []*Comment
	resp, err := service.client.Do(req, &comments)
	if err != nil {
		return nil, resp, err
	}

	return comments, resp, nil
}

// ListBlockers returns the list of Blockers in a Story.
func (service *StoryService) ListBlockers(
	projectId int,
	storyId int,
) ([]*Blocker, *http.Response, error) {

	u := fmt.Sprintf("projects/%v/stories/%v/blockers", projectId, storyId)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var blockers []*Blocker
	resp, err := service.client.Do(req, &blockers)
	if err != nil {
		return nil, resp, err
	}

	return blockers, resp, nil
}

func (service *StoryService) AddBlocker(projectId int, storyId int, description string) (*Blocker, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v/blockers", projectId, storyId)
	req, err := service.client.NewRequest("POST", u, BlockerRequest{
		Description: description,
	})
	if err != nil {
		return nil, nil, err
	}

	var blocker Blocker
	resp, err := service.client.Do(req, &blocker)
	if err != nil {
		return nil, resp, err
	}

	return &blocker, resp, nil
}

func (service *StoryService) UpdateBlocker(projectId, stroyId, blockerId int, blocker *BlockerRequest) (*Blocker, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v/blockers/%v", projectId, stroyId, blockerId)
	req, err := service.client.NewRequest("PUT", u, blocker)
	if err != nil {
		return nil, nil, err
	}

	var blockerResp Blocker
	resp, err := service.client.Do(req, &blockerResp)
	if err != nil {
		return nil, resp, err
	}

	return &blockerResp, resp, nil
}
