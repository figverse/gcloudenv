# gcloudenv shell integration for fish.
# Installed via:  gcloudenv init fish | source

function gcloudenv
	switch "$argv[1]"
		case use
			set -l out (command gcloudenv export --shell fish $argv[2..-1]); or return $status
			for line in $out
				eval $line
			end
		case '*'
			command gcloudenv $argv
	end
end

function _gcloudenv_auto --on-variable PWD
	set -l out (command gcloudenv export --shell fish --auto 2>/dev/null); or return 0
	for line in $out
		eval $line
	end
end

_gcloudenv_auto
