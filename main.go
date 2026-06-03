// Command gcloudenv manages gcloud configurations the way nvm/rbenv manage
// language versions: switch the active profile per-shell, set a global default,
// and auto-switch based on a .gcloudenv file in the working directory.
package main

import "github.com/figverse/gcloudenv/cmd"

func main() {
	cmd.Execute()
}
