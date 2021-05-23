
# { stdenv, lib, fetchgit, buildGoModule, zlib, makeWrapper, xcodeenv, androidenv
# ,
# ,
{
    pkgs ? import <nixpkgs> {}
}:
with pkgs;
mkShell {
    buildInputs = [
        # xcodeWrapper
        (ruby_2_7.withPackages (ps: [
            ps.ffi-compiler
        ]))

        nodejs-16_x
        yarn
        watchman
        go_1_16
        (gomobile.override {
            androidPkgs = androidenv.composeAndroidPackages {
                includeNDK = true;
                ndkVersion = "21.3.6528147"; # WARNING: 22.0.7026061 is broken.
            };
            xcodeWrapperArgs = { version = "12.5"; };
        })
  ] ++ lib.optionals stdenv.isDarwin [
      libffi
      libffi.dev
      cocoapods
      (xcodeenv.composeXcodeWrapper {
          version = "12.5";
          xcodeBaseDir = "/Applications/Xcode.app";
      })
      darwin.apple_sdk.frameworks.CoreFoundation
      darwin.apple_sdk.frameworks.CoreBluetooth
  ] ++ lib.optionals stdenv.isLinux [
    docker
    docker-compose
  ];

  shellHook = ''
    export CGO_LDFLAGS+=" -F/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks"
    export NIX_LDFLAGS_AFTER+=" -F/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/System/Library/Frameworks -L/usr/lib -F/Library/Frameworks -F/System/Library/Frameworks"
  '';
}
