package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/jobq/job"
	"github.com/Ishan27g/go-utils/jobq/task"
	"github.com/heimdalr/dag"
)

type Queue[j job.Job] interface {
	getDag() *dag.DAG

	DefaultTask(j) task.Task
	Add(j) string
	Run() error
	ResetAfter(ids ...string)
}

type queue struct {
	dg *dag.DAG
}

func (q *queue) getDescendants(id string, dg *dag.DAG) (map[string]task.Task, []string, bool) {
	all, _ := dg.GetDescendants(id)
	orderedIds, err := dg.GetOrderedDescendants(id)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, false
	}
	ids := make(map[string]task.Task)
	for _, descendant := range orderedIds {
		if tk, ok := all[descendant].(task.Task); ok {
			ids[tk.Id()] = tk
		}
	}
	return ids, orderedIds, true
}

func (q *queue) Add(j job.Job) string {
	tk := task.New(j, q.getDag())
	return tk.Id()
}

func (q *queue) ResetAfter(ids ...string) {
	// if ids, then reset from an edge
	if len(ids) > 0 && ids[0] != "" {
		for _, id := range ids {
			if roots, orderedIds, ok := q.getDescendants(id, q.dg); ok {
				for _, edges := range orderedIds {
					roots[edges].ResetRun()
				}
			}
		}
		return
	}
	// otherwise, reset from roots
	for _, t := range q.dg.GetRoots() {
		if tk, ok := t.(task.Task); ok {
			tk.ResetRun()
		}
	}
}

func (q *queue) DefaultTask(j job.Job) task.Task {
	return task.New(j, q.getDag())
}

func (q *queue) getDag() *dag.DAG {
	return q.dg
}

func New() Queue[job.Job] {
	return &queue{dg: dag.NewDAG()}
}
func (q *queue) Run() error {
	var err error
	for _, t := range q.dg.GetRoots() {
		if tk, ok := t.(task.Task); ok {
			err = q.run(tk, err)
		}
	}
	return err
}

func newBuf(cap int) *lock {
	return &lock{
		cap: make(chan struct{}, cap),
	}
}

type lock struct {
	cap chan struct{}
}

func (n *lock) Lock() {
	n.cap <- struct{}{}
}

func (n *lock) Unlock() {
	<-n.cap
}
func wtf() {
	c := sync.NewCond(newBuf(3))
	tc := time.NewTicker(30 * time.Millisecond)
	go func() {
		for range tc.C {
			fmt.Println("c.Wait()")
			c.Wait()
			break
		}
		fmt.Println("c.Wait() break")
	}()

	go func() {
		<-tc.C
		c.Signal()
		fmt.Println("c.Signal()ed")
	}()
}
func (q *queue) run(tk task.Task, err error) error {
	roots, orderedIds, ok := q.getDescendants(tk.Id(), q.dg)
	if !ok {
		return errors.New("bad dag")
	}

	// run root
	err = tk.Run()
	if err != nil {
		return nil
	}

	// run edges
	for _, edges := range orderedIds {
		err = roots[edges].Run()
		if err != nil {
			return err
		}
	}
	return err
}
