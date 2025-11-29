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
    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    perSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
  in {
    packages = perSystem (system: let
      callPackage = nixpkgs.darwin.apple_sdk_11_0.callPackage or nixpkgs.legacyPackages.${system}.callPackage;
    in {
      default = callPackage ./. {
        inherit (gomod2nix.legacyPackages.${system}) buildGoApplication;
      };
    });

    devShells = forEachSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
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
              git-chglog
              svu
            ];

            pre-commit.hooks.gomod2nix = {
              enable = true;
              always_run = true;
              name = "gomod2nix";
              description = "Run gomod2nix before commit";
              pass_filenames = false;
              entry = "${gomod2nix.legacyPackages.${system}.gomod2nix}/bin/gomod2nix";
            };
            pre-commit.hooks.changelog = {
              enable = true;
              always_run = true;
              name = "change";
              description = "Generate a changelog";
              pass_filenames = false;
              entry = "nix .#changelog";
            };
          }
        ];
      };
    });
    apps = forEachSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      changelog-script = pkgs.writeShellScriptBin "changelog" ''
        set -euo pipefail
        current_tag=$(${pkgs.svu}/bin/svu current)
        ${pkgs.git-chglog}/bin/git-chglog "$current_tag"
        echo "Generated CHANGELOG.md from $current_tag"
      '';
    in {
      changelog = {
        type = "app";
        program = "${changelog-script}/bin/changelog";
      };
    });
  };
}
