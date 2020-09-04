
# pull-request-code-coverage

[![Build Status](https://vela-server.prod.target.com/badge/TargetOSS/pull-request-code-coverage/status.svg)](https://vela.prod.target.com/TargetOSS/pull-request-code-coverage)

Drone plugin to allow detecting code coverage for only the lines changed in a PR.

Sometimes when working to get a repo to an acceptable level of code coverage, it can be hard to tell if one change is
covered enough.  This plugin will look at just the lines changed in the PR and report code coverage for only those
lines.

This plugin will output the coverage details to the drone step's console, as well as has the ability to comment on the 
PR with a summary of the coverage details.



## VELA Usage

```yaml
- name: check-pr-code-coverage
   image: docker.target.com/app/pull-request-code-coverage
   pull: true
   ruleset:
     event: [pull_request]
   parameters:
     coverage_type: jacoco
     coverage_file: some-sub-module/build/reports/jacoco/test/jacocoTestReport.xml
     source_dirs:
       - src/main/java
       - src/main/kotlin
     gh_api_base_url: https://git.target.com
     module: some-sub-module
   secrets:
     - source: pull_request_api_key
       target: plugin_gh_api_key
```

## DRONE Usage

```yaml
check-pr-code-coverage:
  when:
    event: [pull_request]
  image: docker.target.com/app/pull-request-code-coverage
  pull: true
  coverage_type: jacoco
  coverage_file: some-sub-module/build/reports/jacoco/test/jacocoTestReport.xml
  source_dirs:
    - src/main/java
    - src/main/kotlin
  gh_api_base_url: https://git.target.com
  module: some-sub-module
  secrets:
  - source: pull_request_api_key
    target: plugin_gh_api_key
```



#### Parameters

|param|required| default | description|
|---|---|---|---|
|coverage_type| true | | **supported values**: jacoco, cobertura<br><br>sets the coverage file format  |
|coverage_file| true | | path to where the coverage file will be located, relative to the working dir |
|source_dirs| true | | array of source dirs, relative to the working dir |
|gh_api_base_url| false | | base url of the gh api for posting coverage comments<br><br>if not set, coverage details will not be commented on PR   |
|gh_api_key| false | | api key to auth for posting coverage comments<br><br>if not set, coverage details will not be commented on PR  |
|module | false  | \<empty string\> | sub-module to use if operating inside a multi-module project (e.g. gradle multi-project build) |
