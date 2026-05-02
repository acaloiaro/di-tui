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

    crossTargets = [
      { goos = "linux";  goarch = "amd64"; }
      { goos = "linux";  goarch = "arm64"; }
      { goos = "darwin"; goarch = "amd64"; }
      { goos = "darwin"; goarch = "arm64"; }
    ];
  in {
    packages = perSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
      callPackage = nixpkgs.darwin.apple_sdk_11_0.callPackage or pkgs.callPackage;
      gomod2nixPkgs = gomod2nix.legacyPackages.${system};

      basePackage = callPackage ./. {
        inherit (gomod2nixPkgs) buildGoApplication;
      };

      mkCrossTarget = { goos, goarch }:
        basePackage.overrideAttrs (old: {
          pname = "di-tui-${goos}-${goarch}";
          CGO_ENABLED = "0";
          GOOS = goos;
          GOARCH = goarch;
          doCheck = false;
          doInstallCheck = false;
          dontFixup = true;
          installPhase = ''
            mkdir -p $out/bin
            go build -o $out/bin/di-tui-${goos}-${goarch} .
          '';
        });
    in {
      default = basePackage;
    } // builtins.listToAttrs (map (t: {
      name = "di-tui-${t.goos}-${t.goarch}";
      value = mkCrossTarget t;
    }) crossTargets));

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
          }
        ];
      };
    });
  };
}
