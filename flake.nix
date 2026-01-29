{
  description = "manifold-k8s golang nixos flake";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/master";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };
    in {
      devShell = pkgs.mkShell {
        nativeBuildInputs = [ pkgs.bashInteractive ];
        buildInputs = with pkgs; [
          go
        ];
        shellHook = with pkgs; ''
          # fixes libstdc++ issues and libgl.so issues
          export LD_LIBRARY_PATH=${lib.makeLibraryPath [stdenv.cc.cc]}
        '';
      };
    });
}
