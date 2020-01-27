# job-cleaner

Removes completed jobs to try to prevent hitting the 10k job limit in GKE.

By supplying a namespace (launch the app into a namespace, or replace the namespace variable), specific namespaces can be managed.

## Process

### Jobs

The process inspects all the jobs on completion, and removes all the jobs which are not running. Reason of failure (for now) is ignored, however failed executions are logged.
This process will also delete it's own old jobs, so observability is low. The use of a log aggregator might be required to keep insight in job failures.

The process itself is a kubernetes cronjob, which launches a job to execute its tasks. The inherent problem which this has, is that if the number of ran jobs starts to near 10k/day, is that this process itself will also fail. A higher run frequency (aka multiple times a day), could stretch this process to multitudes of the 10k/day limit.

### Pods

Pods on most k8s systems can run in the 100k+ numbers, however all stopped pods take up disk space, and can clog the overview of pods. The job-cleaner for this reason also cleans pods which have completed.

## Build & run

There is a sample kubernetes cronjob yaml file in the buildAndRun directory. No dockerfile has been provided and no docker image is available for security reasons.

The cronjob file provided has been tested with the application. A memory setting of 300MB seemed sufficient not to hit OOM.
