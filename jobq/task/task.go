package task

import (
	"fmt"
	"sync"

	"github.com/Ishan27g/go-utils/jobq/job"
	"github.com/heimdalr/dag"
)

type Task interface {
	job.Job

	ResetRun() // reset for this and all of its edges
	Id() string
	AddChild(Task Task) Task // adds an edge between this and the supplied task
}

type task struct {
	id string
	sync.Mutex
	ran bool
	r   func() error
	dg  *dag.DAG
}

func (t *task) resetRun() {
	t.Lock()
	defer t.Unlock()
	t.ran = false
}

func (t *task) hasRun() bool {
	return t.ran
}
func (t *task) run() error {
	t.Lock()
	defer t.Unlock()
	t.ran = true
	return t.r()
}
func New(j job.Job, dg *dag.DAG) Task {
	t := task{
		id:    "",
		Mutex: sync.Mutex{},
		ran:   false,
		r:     j.Run,
		dg:    dg,
	}

	id, err := t.dg.AddVertex(&t)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	t.id = id
	return &t
}

func (t *task) Id() string {
	return t.id
}

func (t *task) AddChild(t2 Task) Task {
	_ = t.dg.AddEdge(t.id, t2.Id())
	return t
}

func (t *task) ResetRun() {
	t.resetRun()
	all, orderedIds, ok := t.getDescendants()
	if !ok {
		return
	}
	for _, descendant := range orderedIds {
		tt := all[descendant].(*task)
		tt.ResetRun()
	}
}

func (t *task) Run() error {
	if t.hasRun() {
		return nil
	}
	return t.run()
}

func (t *task) getDescendants() (map[string]interface{}, []string, bool) {
	all, _ := t.dg.GetDescendants(t.id)
	orderedIds, err := t.dg.GetOrderedDescendants(t.id)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, false
	}
	return all, orderedIds, true
}
