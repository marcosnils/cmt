# cmt
Container migration tool for the [Docker Global HackDay #3](https://www.docker.com/community/hackathon?mkt_tok=3RkMMJWWfF9wsRonuqTMZKXonjHpfsX57ugoXqe0lMI/0ER3fOvrPUfGjI4AT8dkI%2BSLDwEYGJlv6SgFQ7LMMaZq1rgMXBk%3D)

https://www.youtube.com/watch?v=pwf0-_cs6U4


## Description

Checkpoint & Restore is still a feature which is not generically available to container users. Certain understanding about how it works is needed and it’s most likely that users get errors when trying to perform CR due to some restrictions or differences between the source and the target host. The purpose of the project is to create an external command line tool that can be either used with docker or runC which helps on the task to live migrate containers between different hosts by performing pre-migration validations and allowing to auto-discover suitable target hosts.

## IMPORTANT!!

This project uses custom patched versions of [CRIU](https://github.com/marcosnils/criu) and [runC](https://github.com/marcosnils/runc/tree/pre_dump) to work. It's important to install these specific versions for CMT to work. CRIU patch has been already proposed to upstream, we hold on runC on the other hand because we needed to implement it fast and we're not sure of any possible impact on the project.

*Update 09/21/15*: CRIU patch as been merged to upstream [here](https://github.com/xemul/criu/commit/e3f900f95429bc0447d8e3cff3cbb2e0a19f8d23). Master version should work with CMT.


## Usage

`go get github.com/marcosnils/cmt`

`cmt --help` should list all possible CMT commands and options

## Authentication

CMT uses ssh-agent authentication when trying to communicate between hosts. Make sure your agent has the corresponding credentials before trying to perform any action.

Instruction about how to setup ssh-agent can be found here: http://sshkeychain.sourceforge.net/mirrors/SSH-with-Keys-HOWTO/SSH-with-Keys-HOWTO-6.html


## Design / performance

CMT was thought to be as portable and lightweight as possible. As it relies on ssh heaviliy for remote communication we also took into account SSH session optimizations and concurrent executions
to speed up the whole process.

It was also designed with the idea to be easily adaptable to any underlying mechanism of C/R. This means that when Docker finally implements C/R natively, CMT can take care of all the necessary
heavy duty to perform container migration (image layer diffs included).

## Hooks

CMT supports 3 kind of hooks. A hook is any command that you provide and that CMT will run when reaching some specific state in the migration process.
The supported hooks are:
  - Pre-restore: which is executed right before restoring the container
on the destination host.
  - Post-restore: which is executed after successfully restoring the
  container on the destination host.
  - Failed-restore: which is executed after a failing to restore the
  container on the destination host.

For example:
```
cmt migrate --hook-pre-restore "echo pre restore" --hook-post-restore "echo post restore" --hook-failed-restore "echo failed restore"
```

There are some very useful scenarios for this. For example in AWS you could use the pre-restore hook to move an Elastic Network Interface so the destination has the same IP address of the source.

## FAQ

### What kind of validations does CMT do?

- Binary existence (runC, criu)
- Binary version matching
- Destination host free memory
- Kernel capabilities to perform c/r (`criu check --ms`) 
- CPU capability problem (http://criu.org/Cpuinfo)


### Can CMT perform TCP live migration without end-user disconnection?

Yes, although all the heavy work is done by CRIU, CMT provides some help when migrating TCP connections to avoid end-user disconnect.
We've accomplished this in AWS using ENI and VPC peering connections.

(*Hope to find the time to demo this soon*)

### Is it necessary to perform validations each time when migrating?

No, validations are performed by default as a security measure, using `--force` flag bypasses them.

### What does pre-dump exactly do?

Please refer to the official CRIU documentation for iterative migration specifics. http://criu.org/Iterative_migration

### What does downtime mean?

Refer to the CRIU documentation for downtime/freeze time. (http://criu.org/)

## TODO

Redo this project as it should be done (tests please!!).

We do have some [issues](https://github.com/marcosnils/cmt/issues) we though about implementing but we couldn't find the time. 


## Special mention to:

- Docker and the community for making us leave our comfort zone and hack on cool stuff. We've learnt a lot these past 4 days.
- Medallia Argentina for hosting the Buenos Aires Docker meetup and being excellent people.
- All OS contributors who can make this happen.
- Ross Boucher (@boucher) for dedicating his personal time to help us answering our annoying questions.
