# Safety mechanisms

Doppelganger has two safety mechanisms aimed at avoiding
unintentional synchronization root deletion. Each of these mechanisms detects an
"irregular" condition during synchronization and halts the synchronization cycle
until the user confirms that the condition is intentional.

The first feature detects complete deletion of the synchronization root on one
side of the connection.

The second feature detects replacement of the synchronization root with a root
of a different type on one side of the connection.

In both cases, the user is required to delete the synchronization root on the
side to which the deletion or replacement should propagate, and then use
`doppelganger resume` to continue synchronization for the session.
