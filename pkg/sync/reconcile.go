package sync

func nonDeletionChangesOnly(changes []*Change) []*Change {
	var result []*Change

	for _, c := range changes {
		if c.New != nil {
			result = append(result, c)
		}
	}

	return result
}

type reconciler struct {
	synchronizationMode SynchronizationMode
	ancestorChanges     []*Change
	alphaChanges        []*Change
	betaChanges         []*Change
	conflicts           []*Conflict
}

func (r *reconciler) reconcile(path string, ancestor, alpha, beta *Entry) {
	if alpha.equalShallow(beta) {
		ancestorContents := ancestor.GetContents()
		alphaContents := alpha.GetContents()
		betaContents := beta.GetContents()

		if !ancestor.equalShallow(alpha) {
			r.ancestorChanges = append(r.ancestorChanges, &Change{
				Path: path,
				New:  alpha.copySlim(),
			})
			ancestorContents = nil
		}

		for name := range nameUnion(ancestorContents, alphaContents, betaContents) {
			r.reconcile(
				pathJoin(path, name),
				ancestorContents[name],
				alphaContents[name],
				betaContents[name],
			)
		}

		return
	}

	switch r.synchronizationMode {
	case SynchronizationMode_SynchronizationModeTwoWaySafe:
		r.handleDisagreementBidirectional(path, ancestor, alpha, beta)
	case SynchronizationMode_SynchronizationModeTwoWayResolved:
		r.handleDisagreementBidirectional(path, ancestor, alpha, beta)
	case SynchronizationMode_SynchronizationModeOneWaySafe:
		r.handleDisagreementUnidirectional(path, ancestor, alpha, beta)
	case SynchronizationMode_SynchronizationModeOneWayReplica:
		r.handleDisagreementUnidirectional(path, ancestor, alpha, beta)
	default:
		panic("unhandled synchronization mode")
	}
}

func (r *reconciler) handleDisagreementBidirectional(path string, ancestor, alpha, beta *Entry) {
	alphaDelta := diff(path, ancestor, alpha)
	if len(alphaDelta) == 0 {
		r.alphaChanges = append(r.alphaChanges, &Change{
			Path: path,
			Old:  ancestor,
			New:  beta,
		})
		return
	}
	betaDelta := diff(path, ancestor, beta)
	if len(betaDelta) == 0 {
		r.betaChanges = append(r.betaChanges, &Change{
			Path: path,
			Old:  ancestor,
			New:  alpha,
		})
		return
	}

	if r.synchronizationMode == SynchronizationMode_SynchronizationModeTwoWayResolved {
		r.betaChanges = append(r.betaChanges, &Change{
			Path: path,
			Old:  beta,
			New:  alpha,
		})
		return
	}

	alphaDeltaNonDeletion := nonDeletionChangesOnly(alphaDelta)
	betaDeltaNonDeletion := nonDeletionChangesOnly(betaDelta)
	if len(alphaDeltaNonDeletion) == 0 {
		r.alphaChanges = append(r.alphaChanges, &Change{
			Path: path,
			Old:  alpha,
			New:  beta,
		})
		return
	} else if len(betaDeltaNonDeletion) == 0 {
		r.betaChanges = append(r.betaChanges, &Change{
			Path: path,
			Old:  beta,
			New:  alpha,
		})
		return
	}

	r.conflicts = append(r.conflicts, &Conflict{
		AlphaChanges: alphaDeltaNonDeletion,
		BetaChanges:  betaDeltaNonDeletion,
	})
}

func (r *reconciler) handleDisagreementUnidirectional(path string, ancestor, alpha, beta *Entry) {
	if r.synchronizationMode == SynchronizationMode_SynchronizationModeOneWayReplica {
		r.betaChanges = append(r.betaChanges, &Change{
			Path: path,
			Old:  beta,
			New:  alpha,
		})
		return
	}

	betaDeltaNonDeletion := nonDeletionChangesOnly(diff(path, ancestor, beta))
	if len(betaDeltaNonDeletion) == 0 {
		r.betaChanges = append(r.betaChanges, &Change{
			Path: path,
			Old:  beta,
			New:  alpha,
		})
		return
	}

	ancestorOrBetaNonDirectory := ancestor == nil ||
		ancestor.Kind != EntryKind_Directory ||
		beta == nil ||
		beta.Kind != EntryKind_Directory
	if alpha == nil && ancestorOrBetaNonDirectory {
		if ancestor != nil {
			r.ancestorChanges = append(r.ancestorChanges, &Change{Path: path})
		}
		return
	}

	alphaDelta := diff(path, ancestor, alpha)
	if len(alphaDelta) == 0 {
		alphaDelta = []*Change{{Path: path, Old: alpha, New: alpha}}
	}
	r.conflicts = append(r.conflicts, &Conflict{
		AlphaChanges: alphaDelta,
		BetaChanges:  betaDeltaNonDeletion,
	})
}

func Reconcile(
	ancestor, alpha, beta *Entry,
	synchronizationMode SynchronizationMode,
) ([]*Change, []*Change, []*Change, []*Conflict) {
	r := &reconciler{
		synchronizationMode: synchronizationMode,
	}

	r.reconcile("", ancestor, alpha, beta)

	return r.ancestorChanges, r.alphaChanges, r.betaChanges, r.conflicts
}
