{ lib
, buildGoModule
}:

buildGoModule rec {
  pname = "launchit";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-lqPdlKDgRGcRwY1VcECwz0ZujadCbFP0NfpndwDKUy4=";

  # The module is at the root, but the main package is in cmd/
  subPackages = [ "cmd" ];

  # Rename the binary from 'cmd' to 'launchit'
  postInstall = ''
    mv $out/bin/cmd $out/bin/launchit
  '';

  meta = with lib; {
    description = "A launcher tool for managing application entries";
    homepage = "https://github.com/jplein/launchit";
    license = licenses.mit; # Update this to match your actual license
    maintainers = [ ];
    mainProgram = "launchit";
  };
}
