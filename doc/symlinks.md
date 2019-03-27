# Symlinks

Doppelganger has full support for symbolic links, both on POSIX and Windows systems.
- **Ignore**: In this mode, Doppelganger simply ignores any symlinks.

- **Portable** (default): In this mode, Doppelganger restricts itself to 
  portable symlinks, which are those that have relative paths.
  . 
- **POSIX**: In this mode, which is only supported for synchronization
   targets without any analysis or modification.

These modes can be specified on a per-session basis by passing the
`--symlinkk-mode=<mode>` flag to the `create` command and by a  
 default basis by including the following configuration in `~/.doppelganger.toml`:

    [symlink]
    mode = "<mode>"



