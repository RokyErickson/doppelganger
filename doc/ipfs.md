# InterPlantaryFileSystem
**IPFS** is an *Experimental* Feature
Doppelganger provides support for synchronizing with filesystem locations accessible via IPFS.
Doppelganger requires an IPFS client installation to be available on your system **AND** the *ipfs-exec.sh* script. 
IPFS synchronization endpoints can be specified to Doppelganger's `create` command
using URL syntax: 

	ipfs:path 
	
The `path` component must be a absoulte Interplantary FileSystem Path:

	/ipfs/<path>
	
	
The ipfs runtime is dictated by one enviroment variable: 
	
- `IPFS_PATH`

Doppelganger is aware of these environment variables and will lock them in at session creation time.

*Note* IPFS is supported bare minimum. IPFS is based on a merkel dag and sync CID's are immutable it is not possible to contiously sync
in the traditonal sense like doppelganger normally does. Right now we simply download content from ipfs and exec a doppelganger endpoint
into the output directory and sync to whatever other endpoint we have. In the future, I would like doppelganger to be able to manage the metadata
of the merkle dag in a manner similar to git that would allow synchnization and version control over ipfs in a much more robust way similar
to dropbox, but that's about 50,000-100,000 lines away.
