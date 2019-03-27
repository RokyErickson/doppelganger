# MOSH
Doppelganger provides support for synchronizing with filesystem locations accessible via MOSH.
Mosh is a more robust solution to, but still similar to SSH and uses ssh for login.
Doppelganger requires an OpenMOSH client installation to be available on your system.
MOSH synchronization endpoints can be specified to Doppelganger's `create` command
using URL syntax:

    mosh:[user@]host[:port]:path

The `user` component is optional. 

The `host` component can be any IP address, hostname, or alias understood by OpenMOSH.

The `port` component is also optional.

The `path` component: 

a absoulte path
	
	mosh:user@host:/var/www

a relative path 

    mosh:user@host:path/in/home/directory

a home-directory-relative path, e.g.

    mosh:user@host:~/path/in/home/directory
-----------------------------------------

*Note* Mosh support is a work in progress and is not functional just yet. Needs in order for the endpoint to connect correctly. Still figuring it out.
