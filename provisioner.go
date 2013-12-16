package urknall

// Provision the given list of runlists.
func provisionRunlists(runLists []*Runlist, provisionFunc func(*Runlist) error) (e error) {
	for i := range runLists {
		if e = provisionFunc(runLists[i]); e != nil {
			Publish("runlists.provision.error", e.Error())
			return e
		}
	}
	return nil
}
