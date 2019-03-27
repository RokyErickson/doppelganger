package sync

type differ struct {
	changes []*Change
}

func (d *differ) diff(path string, base, target *Entry) {

	if !target.equalShallow(base) {
		d.changes = append(d.changes, &Change{
			Path: path,
			Old:  base,
			New:  target,
		})
		return
	}

	baseContents := base.GetContents()
	targetContents := target.GetContents()
	for name := range nameUnion(baseContents, targetContents) {
		d.diff(pathJoin(path, name), baseContents[name], targetContents[name])
	}
}

func diff(path string, base, target *Entry) []*Change {

	d := &differ{}

	d.diff(path, base, target)

	return d.changes
}
