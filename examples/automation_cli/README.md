<!--
 Copyright 2023 Searis AS

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Automation CLI

This example shows how to use the [automationcli][cli] package to build your own CLI tool for running automation routines from the command-line. This is a basic example. For setting up your production routines, and running them on a schedule using GitHub actions or other means, you can fork the [template-clarify-automation][template] repo to get started quickly.

[cli]: https://pkg.go.dev/github.com/clarify/clarify-go/automation/automationcli
[template]: https://github.com/clarify/template-clarify-automation

The example contains the following automation routines:

- devdata/save-signals: Update meta-data for a status signal in Clarify. No requirements.
- devdata/insert-random: Insert a state for the status signal. No requirements.
- publish: Insert a state for the status signal. Require setting `CLARIFY_EXAMPLE_PUBLISH_INTEGRATION_ID`.
- evaluate/detect-fire: Log a message to the console if a fire is detected.

## Running the example

First, you myst download a credentials file for your integration, and a configure the integration to have access to all namespaces. This require that you are an admin in your clarify organization. To avoid passing in options each time, we will export it as a environment variable:

```sh
export CLARIFY_CREDENTIALS=path/to/clarify-credentials.json
```

Secondly, load the the database using the devdata routines. Note that the routines are run in an alphanumerical order.

```sh
go run . -v devdata
```

Now, let's publish the signals. For that we need to know the integration ID we should publish from. You can find the integration ID in the Admin panel. Ideally, we would edit the routine code to hard-code the integration IDs to publish from. However, for this example, we allow parsing the information from the environment:

```sh
export CLARIFY_EXAMPLE_PUBLISH_INTEGRATION_ID=...
go run . -v publish
```

Now, let's run our evaluation routine. For that, we need to know the ID from the item that was just published. You can get this ID from the console output, or from Clarify.

```sh
export CLARIFY_EXAMPLE_STATUS_ITEM_ID=...
go run . evaluate/detect-fire
```

You should now have about 45 % chance to see the text "FIRE! FIRE! FIRE!".
