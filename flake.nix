{
  description = "A beautiful TUI for managing Claude Code sessions";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = self.shortRev or "dirty";
      in
      {
        packages.default = pkgs.buildGo125Module {
          inherit version;

          pname = "faustus";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-RNbS40G+8rtwlSJgYLN1puTCytGfXdagQTEs6sIXwnM=";
          ldflags = [
            "-s"
            "-w"
          ];

          meta = with pkgs.lib; {
            description = "A beautiful TUI for managing Claude Code sessions";
            homepage = "https://github.com/Fuwn/faustus";
            license = licenses.gpl3Only;
            platforms = platforms.unix;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [ pkgs.go_1_24 ];
        };
      }
    );
}
