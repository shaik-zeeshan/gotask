package queue

import (
	"encoding/json"
	"log"
)

type JobType map[string]func(string)

type Jobs struct {
	Jobs JobType
}

type StatusJob string

const (
	Waiting   StatusJob = "waiting"
	Completed           = "completed"
	Running             = "running"
	Failed              = "failed"
)

type Job struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Payload string    `json:"payload"`
	Type    string    `json:"type"`
	Status  StatusJob `json:"status"`
}

func NewJob(ID int, Name string, Payload interface{}, Type string) *Job {
	jsonPayload, err := json.Marshal(Payload)
	if err != nil {
		return nil
	}
	return &Job{
		ID:      ID,
		Name:    Name,
		Type:    Type,
		Payload: string(jsonPayload),
	}
}

func (js *Jobs) SelectJob(j Job) {
	if function, ok := js.Jobs[j.Type]; ok {
		log.Println("here")
		function(j.Payload)
	} else {
		panic("Unknown Job Type")
	}
}

var StoreJob []*Job

var dataChan = make(chan Job)

func HandleJobs(jobs Jobs) {
	for n := range dataChan {
		StoreJob = append(StoreJob, &n)
		go func() {
			n.Status = Running
			defer func() {
				if r := recover(); r == nil {
					n.Status = Completed
				} else {
					log.Println(r)
					n.Status = Failed
				}
			}()
			jobs.SelectJob(n)
		}()
	}
}

func CreateNewJob[T interface{}](Name string, Payload T, Type string) {
	dataChan <- *NewJob(len(StoreJob)+1, Name, Payload, Type)
}

func GetAllJobs() *[]*Job {
	return &StoreJob
}
