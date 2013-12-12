package zwo

// Provision the given list of runlists.
func provisionRunlists(runLists []*Runlist, provisionFunc func(*Runlist) error) (e error) {
	for i := range runLists {
		rl := runLists[i]

		logger.PushPrefix(padToFixedLength(rl.name, 15))

		if e = provisionFunc(runLists[i]); e != nil {
			logger.Errorf("failed to provision: %s", e.Error())
			return e
		}

		logger.PopPrefix()
	}
	return nil
}
