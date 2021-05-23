{
  description = "Berty environment";

  inputs = {
      # channels
      nixpkgs = { url = "github:nixos/nixpkgs/nixos-unstable"; };
      # nixpkgs-master = { url = "github:nixos/nixpkgs/master"; };
      # nixpkgs-stable-darwin = { url = "github:nixos/nixpkgs/nixpkgs-20.09-darwin"; };
      # nixpkgs = { url = "github:nixos/nixpkgs/staging-next"; };

      # android
      android-nixpkgs = {
          inputs.nixpkgs.follows = "nixpkgs";
          url = "github:tadfisher/android-nixpkgs";
      };

      # flake
      flake-utils = { url = "github:numtide/flake-utils"; };
  };

  outputs = { self, nixpkgs, android-nixpkgs, flake-utils }:
      flake-utils.lib.eachDefaultSystem
          (system:
              let pkgs = import nixpkgs {
                      system = system;
                      config = {
                          android_sdk.accept_license = true;
                      };
                  };
              in
              {
                  devShell = import ./shell.nix { inherit pkgs; };
              }
          );
}
