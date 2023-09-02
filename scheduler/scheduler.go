package scheduler

type Scheduler interface {
	Pick()
	SelectCandidateNodes()
	Score()
}
