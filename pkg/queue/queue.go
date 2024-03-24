package queue

import "errors"

type Queue struct {
    items []interface{}
}

func (q *Queue) Enque(item interface{}) { 
    q.items = append(q.items, item)
}

func (q *Queue) Deque() (interface{}, error) {
    if len(q.items) == 0 {
        return nil, errors.New("Can not deque from empty queue")
    }

    var item interface{} = q.items[0]
    q.items = q.items[1:]
    return item, nil
}

func (q *Queue) Size() int {
    return len(q.items)
}
