// Copyright (c) 2018 Salsita Software
// Use of this source code is governed by the MIT License.
// The license can be found in the LICENSE file.

package pivotal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Epic is the primary data object for the epic service.
type Epic struct {
	ID          int        `json:"id,omitempty"`
	ProjectID   int        `json:"project_id,omitempty"`
	Name        string     `json:"name,omitempty"`
	LabelID     int        `json:"label_id,omitempty"`
	Description string     `json:"description,omitempty"`
	CommentIDs  []int      `json:"comment_ids,omitempty"`
	FollowerIDs []int      `json:"follower_ids,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	AfterID     int        `json:"after_id,omitempty"`
	BeforeID    int        `json:"before_id,omitempty"`
	URL         string     `json:"url,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Kind        string     `json:"kind,omitempty"`
}

// EpicRequest is used to get the Epics back.
type EpicRequest struct {
	ProjectID   int       `json:"project_id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Label       Label     `json:"label,omitempty"`
	LabelID     int       `json:"label_id,omitempty"`
	Description string    `json:"description,omitempty"`
	Comments    []Comment `json:"comments,omitempty"`
	CommentIDs  []int     `json:"comment_ids,omitempty"`
	Followers   []Person  `json:"followers,omitempty"`
	FollowerIDs []int     `json:"follower_ids,omitempty"`
	AfterID     int       `json:"after_id,omitempty"`
	BeforeID    int       `json:"before_id,omitempty"`
}

// EpicService wraps the client context to do actions.
type EpicService struct {
	client *Client
}

func newEpicService(client *Client) *EpicService {
	return &EpicService{client}
}

// List returns all epics matching the filter in case the filter is specified.
//
// List actually sends 2 HTTP requests - one to get the total number of epics,
// another to retrieve the epics using the right pagination setup. The reason
// for this is that the filter might require to fetch all the epics at once
// to get the right results. Since the response as generated by Pivotal Tracker
// is not always sorted when using a filter, this approach is required to get
// the right data. Not sure whether this is a bug or a feature.
func (service *EpicService) List(projectID int, filter string) ([]*Epic, error) {
	reqFunc := newEpicsRequestFunc(service.client, projectID, filter)
	cursor, err := newCursor(service.client, reqFunc, 0)
	if err != nil {
		return nil, err
	}

	var epics []*Epic
	if err := cursor.all(&epics); err != nil {
		return nil, err
	}
	return epics, nil
}

func newEpicsRequestFunc(client *Client, projectID int, filter string) func() *http.Request {
	return func() *http.Request {
		u := fmt.Sprintf("projects/%v/epics", projectID)
		if filter != "" {
			u += "?filter=" + url.QueryEscape(filter)
		}
		req, _ := client.NewRequest("GET", u, nil)
		return req
	}
}

// EpicCursor is used to implement the iterator pattern.
type EpicCursor struct {
	*cursor
	buff []*Epic
}

// Next returns the next epic.
//
// In case there are no more epics, io.EOF is returned as an error.
func (c *EpicCursor) Next() (e *Epic, err error) {
	if len(c.buff) == 0 {
		_, err = c.next(&c.buff)
		if err != nil {
			return nil, err
		}
	}

	if len(c.buff) == 0 {
		err = io.EOF
	} else {
		e, c.buff = c.buff[0], c.buff[1:]
	}
	return e, err
}

// Iterate returns a cursor that can be used to iterate over the epics specified
// by the filter. More epics are fetched on demand as needed.
func (service *EpicService) Iterate(projectID int, filter string) (c *EpicCursor, err error) {
	reqFunc := newEpicsRequestFunc(service.client, projectID, filter)
	cursor, err := newCursor(service.client, reqFunc, PageLimit)
	if err != nil {
		return nil, err
	}
	return &EpicCursor{cursor, make([]*Epic, 0)}, nil
}

// Create is used to create a new Epic with an EpicRequest.
func (service *EpicService) Create(projectID int, epic *EpicRequest) (*Epic, *http.Response, error) {
	if projectID == 0 {
		return nil, nil, &ErrFieldNotSet{"project_id"}
	}

	if epic.Name == "" {
		return nil, nil, &ErrFieldNotSet{"name"}
	}

	u := fmt.Sprintf("projects/%v/epics", projectID)
	req, err := service.client.NewRequest("POST", u, epic)
	if err != nil {
		return nil, nil, err
	}

	var newEpic Epic

	resp, err := service.client.Do(req, &newEpic)
	if err != nil {
		return nil, resp, err
	}

	return &newEpic, resp, nil
}

// Get is returns an Epic by ID.
func (service *EpicService) Get(projectID, epicID int) (*Epic, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/epics/%v", projectID, epicID)
	req, err := service.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var epic Epic
	resp, err := service.client.Do(req, &epic)
	if err != nil {
		return nil, resp, err
	}

	return &epic, resp, err
}

// Update is will update an Epic with an EpicRequest.
func (service *EpicService) Update(projectID, epicID int, epic *EpicRequest) (*Epic, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v", projectID, epicID)
	req, err := service.client.NewRequest("PUT", u, epic)
	if err != nil {
		return nil, nil, err
	}

	var updatedEpic Epic
	resp, err := service.client.Do(req, &updatedEpic)
	if err != nil {
		return nil, resp, err
	}

	return &updatedEpic, resp, err

}
