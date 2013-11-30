package projects

type Job func()
type JobControlCommand int

type ConcurrentProcessor struct {
	PendingJobs       int
	ControlChannel    chan JobControlCommand
	JobChannel        chan Job
	CompletionChannel chan bool
}

const (
	IncrementPendingJobs JobControlCommand = iota
	CompletePendingJob
)

func NewConcurrentProcessor() *ConcurrentProcessor {
	control := make(chan JobControlCommand, 10000)
	jobs := make(chan Job, 10000)
	completion := make(chan bool)
	return &ConcurrentProcessor{0, control, jobs, completion}
}

func (processor *ConcurrentProcessor) WaitForCompletion() {
	<-processor.CompletionChannel
}

func (processor *ConcurrentProcessor) AddJob(job Job) {
	processor.ControlChannel <- IncrementPendingJobs
	processor.JobChannel <- job
}

func (processor *ConcurrentProcessor) StartManager() {
	go processor.RunManager()
}

func (processor *ConcurrentProcessor) StartWorkers(workers int) {
	for i := 0; i < workers; i++ {
		go processor.SpawnWorker()
	}
}

func (processor *ConcurrentProcessor) RunManager() {
	for i := range processor.ControlChannel {
		switch i {
		case IncrementPendingJobs:
			processor.PendingJobs++
		case CompletePendingJob:
			processor.PendingJobs--
			if processor.PendingJobs <= 0 {
				close(processor.JobChannel)
				processor.CompletionChannel <- true
			}
		}
	}

}

func (processor *ConcurrentProcessor) SpawnWorker() {
	for job := range processor.JobChannel {
		job()
		processor.ControlChannel <- CompletePendingJob
	}
}
