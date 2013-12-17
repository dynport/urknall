package urknall

// Provision the given list of runlists.
func provisionRunlists(runLists []*Runlist, provisionFunc func(*Runlist) error) (e error) {
	for i := range runLists {
		rl := runLists[i]
		m := &Message{key: MessageRunlistsProvision, runlist: rl}
		m.publish("started")
		if e = provisionFunc(rl); e != nil {
			m.publishError(e)
			return e
		}
		m.publish("finished")
	}
	return nil
}
