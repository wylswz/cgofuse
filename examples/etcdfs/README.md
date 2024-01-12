# Etcd FS

This is a file system backed by etcd. It works when keys in etcd are organized in filestystem path style.


## Goals
- Read etcd key/values
- Access keys like file system directories
- Edit values in etcd

## Non-goals
- Locking
- Links
- Support any other file types except regular files and directories
- Fine-grained access control

## Common Engine Interfae

Implement the `Engine` interface in `pkg/engine` in order to support other database types



# VS code configuration

```json
{
    "version": "0.2.0",
    "configurations": [
        
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "dlvFlags": ["--only-same-user=false"],
            "args": ["/home/makaveli/etcdfs", "-d"]
        }
    ]
}
```