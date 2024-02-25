{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    devenv.url = "github:cachix/devenv";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    devenv,
    systems,
    gomod2nix,
    ...
  } @ inputs: let
    forEachSystem = nixpkgs.lib.genAttrs (import systems);
  in {
    devShells = forEachSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      # The current default sdk for macOS fails to compile go projects, so we use a newer one for now.
      # This has no effect on other platforms.
      callPackage = pkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
    in {
      packages.default = callPackage ./. {
        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
      };

      default = devenv.lib.mkShell {
        inherit inputs pkgs;
        modules = [
          {
            languages.go = {
              enable = true;
              package = pkgs.go_1_22;
            };

            packages = with pkgs; [
              gomod2nix.legacyPackages.${system}.gomod2nix
              golangci-lint
              pre-commit
            ];

            pre-commit.hooks.gomod2nix = {
              enable = true;
              always_run = true;
              name = "gomod2nix";
              description = "Run gomod2nix before commit";
              pass_filenames = false;
              entry = "${gomod2nix.legacyPackages.${system}.gomod2nix}/bin/gomod2nix";
            };
          }
        ];
      };
    });
  };
}
