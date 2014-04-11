package pulse

import (
	"errors"
	"strconv"
	"time"

	"github.com/kolo/xmlrpc"
)

// Client TODO(rjeczalik): document
type Client interface {
	Agents() (Agents, error)
	BuildID(reqid string) (int64, error)
	BuildResult(project string, id int64) ([]BuildResult, error)
	Clear(project string) error
	Close() error
	Init(project string) (bool, error)
	LatestBuildResult(project string) ([]BuildResult, error)
	Projects() ([]string, error)
	Stages(project string) ([]string, error)
	Trigger(project string) ([]string, error)
	WaitBuild(project string, id int64) <-chan struct{}
}

type client struct {
	rpc *xmlrpc.Client
	tok string
}

// NewClient TODO(rjeczalik): document
func NewClient(url, user, pass string) (Client, error) {
	c, err := &client{}, (error)(nil)
	if c.rpc, err = xmlrpc.NewClient(url, nil); err != nil {
		return nil, err
	}
	if err = c.rpc.Call("RemoteApi.login", []interface{}{user, pass}, &c.tok); err != nil {
		return nil, err
	}
	return c, nil
}

// Init TODO(rjeczalik): document
func (c *client) Init(project string) (ok bool, err error) {
	err = c.rpc.Call("RemoteApi.initialiseProject", []interface{}{c.tok, project}, &ok)
	return
}

// BuildID TODO(rjeczalik): document
func (c *client) BuildID(reqid string) (int64, error) {
	// TODO(rjeczalik): Extend Client interface with SetDeadline() method which
	//                  will configure both timeouts - for Pulse API and for
	//                  *rpc.Client from net/rpc.
	timeout, rep := 15*1000, &BuildRequestStatus{}
	err := c.rpc.Call("RemoteApi.waitForBuildRequestToBeActivated",
		[]interface{}{c.tok, reqid, timeout}, &rep)
	if err != nil {
		return 0, err
	}
	if rep.Status == BuildUnhandled || rep.Status == BuildQueued {
		return 0, errors.New("pulse: requesting build ID has timed out")
	}
	return strconv.ParseInt(rep.ID, 10, 64)
}

// BuildResult TODO(rjeczalik): document
func (c *client) BuildResult(project string, id int64) ([]BuildResult, error) {
	var build []BuildResult
	err := c.rpc.Call("RemoteApi.getBuild", []interface{}{c.tok, project, int(id)}, &build)
	if err != nil {
		return nil, err
	}
	return build, nil
}

// Stages TODO(rjeczalik): document
func (c *client) Stages(project string) ([]string, error) {
	// TODO(rjeczalik): It would be better to get stages list from project's configuration.
	//                  I ran away screaming while trying to get that information from
	//                  the Remote API spec.
	b, err := c.LatestBuildResult(project)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("pulse: error requesting latest build status")
	}
	if len(b[0].Stages) == 0 {
		return nil, errors.New("pulse: stage list is empty")
	}
	s := make([]string, 0, len(b[0].Stages))
	for i := range b[0].Stages {
		s = append(s, b[0].Stages[i].Name)
	}
	return s, nil
}

// LatestBuildResult TODO(rjeczalik): document
func (c *client) LatestBuildResult(project string) ([]BuildResult, error) {
	var build []BuildResult
	err := c.rpc.Call("RemoteApi.getLatestBuildForProject", []interface{}{c.tok, project, true}, &build)
	if err != nil {
		return nil, err
	}
	return build, nil
}

// WaitBuild TODO(rjeczalik): document
func (c *client) WaitBuild(project string, id int64) <-chan struct{} {
	done, sleep := make(chan struct{}), 250*time.Millisecond
	go func() {
		build, retry := make([]BuildResult, 0), 3
	WaitLoop:
		for {
			build = build[:0]
			err := c.rpc.Call("RemoteApi.getBuild", []interface{}{c.tok, project,
				int(id)}, &build)
			if err != nil {
				if retry > 0 {
					retry -= 1
					time.Sleep(sleep)
					continue WaitLoop
				}
				close(done)
				return
			}
			for i := range build {
				if !build[i].Complete {
					time.Sleep(sleep)
					continue WaitLoop
				}
			}
			close(done)
			return
		}
	}()
	return done
}

// Close TODO(rjeczalik): document
func (c *client) Close() error {
	if err := c.rpc.Call("RemoteApi.logout", c.tok, nil); err != nil {
		return err
	}
	return c.rpc.Close()
}

// Clear TODO(rjeczalik): document
func (c *client) Clear(project string) error {
	return c.rpc.Call("RemoteApi.doConfigAction", []interface{}{c.tok, "projects/" + project, "clean"}, nil)
}

// Trigger TODO(rjeczalik): document
func (c *client) Trigger(project string) (id []string, err error) {
	// TODO(rjeczalik): Use TriggerOptions struct instead after kolo/xmlrpc
	//                  supports maps.
	req := struct {
		R bool `xmlrpc:"rebuild"`
	}{true}
	err = c.rpc.Call("RemoteApi.triggerBuild", []interface{}{c.tok, project, req}, &id)
	return
}

// Projects TODO(rjeczalik): document
func (c *client) Projects() (s []string, err error) {
	err = c.rpc.Call("RemoteApi.getAllProjectNames", c.tok, &s)
	return
}

// Agents TODO(rjeczalik): document
func (c *client) Agents() (Agents, error) {
	var names []string
	if err := c.rpc.Call("RemoteApi.getAllAgentNames", c.tok, &names); err != nil {
		return nil, err
	}
	a := make(Agents, len(names))
	for i := range names {
		if err := c.rpc.Call("RemoteApi.getAgentDetails", []interface{}{c.tok, names[i]}, &a[i]); err != nil {
			return nil, err
		}
		a[i].Name = names[i]
	}
	return a, nil
}
