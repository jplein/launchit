.PHONY: launchit

launchit: clean
	nix build && ln -s ./result/bin/launchit .

clean:
	rm ./result
	rm ./launchit
