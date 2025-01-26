{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    delve
    gdlv
    cobra-cli
  ];

  shellHook = ''
    echo "Welcome to your new go shell"
  '';
}
