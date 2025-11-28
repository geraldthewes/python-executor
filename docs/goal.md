I want to build a remote python execution environment to run in nomad.

I want two API for it, one more complex that consists of the following parts

* Docker image to use (override a default very slim python 3.12 image)
* A set of CLI commands to run before executing the python scripts
* Python packages to install before executing the python script as a requirements.txt
* An optional stdin input
* A configuration file
    This would include for now, but should be extensible
	   - maximum execution time
	   - allow/deny network access
	   - Max memory
	   - Max disk
* One or more python files
* The main python file to be executed

A simpler API will also be developed that just includes
* A requirements.txt of packages to insatll
* The pythin script to execute as a single file
  Note this can simply be accomplished by making all the other items optional

The server component will take such a request, spin up a docker image inside the docker image, execute the prep tasks, then execute the code.

The output of stdout and stderr will be collected and returned to the caller.

include both a synchroneous version
And an async version
   - Caller gets an execution id back
   - Caller can poll
   - Upon completion an entry in consul is updated (easier to poll)
   - Ability to kill execution
   
The output consists of the followmg  deliverables
  * The server as a docker image to be deployed by nomad
  * A documentation of the API (or better the API is self documented swagger style)
  * A python and golang client package to perform the requests to the server
  * A golang command line CLI to make the requests
  
example simplified command is

   cat main.py | python-executor
   
   This syns synchoneously the main.py script
   
Please create a PRD for this project following the PRD template here
https://github.com/snarktank/ai-dev-tasks/blob/main/create-prd.md or
the example here https://github.com/geraldthewes/nomad-mcp-builder/blob/main/prds/PRD.md
   
 Please develop an implementation for the projet described in @docs/PRD.md
  Feel free to learn from
  https://github.com/unclefomotw/code-executor
  https://github.com/geraldthewes/nomad-mcp-builder
  Please ask any questions you may have on this project
  Use only popular, well recommened and supported dependencies
  Use the latest stable version of all dependencies 
───────────────────────────────────────────────────────────────────
