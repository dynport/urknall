package urknall

type checksumTree map[string]map[string]struct{}

type Provisioner interface {
	ProvisionRunlist(*Package, checksumTree) error
	BuildChecksumTree() (checksumTree, error)
}

// Provision the given list of runlists.
func provisionRunlists(runLists []*Package, runner *Runner) (e error) {
	ct, e := buildChecksumTree(runner)
	if e != nil {
		return e
	}

	for i := range runLists {
		rl := runLists[i]
		m := &Message{key: MessageRunlistsProvision, runlist: rl}
		m.publish("started")
		if e = provisionRunlist(runner, rl, ct); e != nil {
			m.publishError(e)
			return e
		}
		m.publish("finished")
	}
	return nil
}
