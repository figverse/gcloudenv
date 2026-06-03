# gcloudenv shell integration for bash/zsh.
# Installed via:  eval "$(gcloudenv init bash)"
#
# The binary cannot change this shell's environment, so `use` is a function
# that evals the export lines the binary prints.

gcloudenv() {
	case "$1" in
	use)
		shift
		local out
		out="$(command gcloudenv export "$@")" || return $?
		eval "$out"
		;;
	*)
		command gcloudenv "$@"
		;;
	esac
}

# Auto-switch when entering a directory that has a .gcloudenv file (opt-in:
# only active once this snippet is sourced). Mirrors nvm's cd behaviour.
_gcloudenv_auto() {
	local out
	out="$(command gcloudenv export --auto 2>/dev/null)" || return 0
	[ -n "$out" ] && eval "$out"
}

if [ -n "$ZSH_VERSION" ]; then
	autoload -U add-zsh-hook 2>/dev/null && add-zsh-hook chpwd _gcloudenv_auto
else
	case "$PROMPT_COMMAND" in
	*_gcloudenv_auto*) ;;
	*) PROMPT_COMMAND="_gcloudenv_auto${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
	esac
fi

_gcloudenv_auto
