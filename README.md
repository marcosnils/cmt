# cmt
Container migration tool for the [Docker Global HackDay #3](https://www.docker.com/community/hackathon?mkt_tok=3RkMMJWWfF9wsRonuqTMZKXonjHpfsX57ugoXqe0lMI/0ER3fOvrPUfGjI4AT8dkI%2BSLDwEYGJlv6SgFQ7LMMaZq1rgMXBk%3D)


## Description

Checkpoint & Restore is still a feature which is not generically available to container users. Certain understanding about how it works is needed and it’s most likely that users get errors when trying to perform CR due to some restrictions or differences between the source and the target host. The purpose of the project is to create an external command line tool that can be either used with docker or runC which helps on the task to live migrate containers between different hosts by performing pre-migration validations and allowing to auto-discover suitable target hosts.

