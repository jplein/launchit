.PHONY: launchit

launchit: 
	nix build && ln -s ./result/bin/launchit .
