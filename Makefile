# Voysnap — local offline voice-to-text dictation for macOS.
#
# Common flow:
#   make whisper   # one-time: clone + build whisper.cpp static libs (Metal)
#   make model     # one-time: download the default speech model
#   make app       # build the Go binary, assemble Voysnap.app, ad-hoc codesign
#   make run       # launch the app (menu-bar icon appears)

APP_NAME    := Voysnap
BUNDLE_ID   := com.kzkr.voysnap
EXECUTABLE  := voysnap

DIST        := dist
APP_DIR     := $(DIST)/$(APP_NAME).app
CONTENTS    := $(APP_DIR)/Contents
MACOS_DIR   := $(CONTENTS)/MacOS
RES_DIR     := $(CONTENTS)/Resources

# whisper.cpp is vendored under third_party and built into static libs.
WHISPER_DIR    := third_party/whisper.cpp
WHISPER_TAG    := v1.7.6
WHISPER_BUILD  := $(WHISPER_DIR)/build
WHISPER_LIB    := $(WHISPER_BUILD)/src/libwhisper.a

# The model is large and not committed; it is downloaded on demand into the
# app's support dir (where the app's default config looks for it).
APP_SUPPORT := $(HOME)/Library/Application Support/$(APP_NAME)
MODELS_DIR  := $(APP_SUPPORT)/models
MODEL_NAME  := ggml-large-v3-turbo.bin
MODEL_URL   := https://huggingface.co/ggerganov/whisper.cpp/resolve/main/$(MODEL_NAME)
MODEL_PATH  := $(MODELS_DIR)/$(MODEL_NAME)

.PHONY: all app build bundle sign signing-identity run install icon whisper model clean clean-app

all: app

## whisper: clone (pinned) and build whisper.cpp static libraries with Metal.
whisper: $(WHISPER_LIB)

$(WHISPER_LIB):
	@if [ ! -d "$(WHISPER_DIR)/.git" ]; then \
		echo ">> cloning whisper.cpp $(WHISPER_TAG)"; \
		git clone --depth 1 --branch $(WHISPER_TAG) https://github.com/ggerganov/whisper.cpp $(WHISPER_DIR); \
	fi
	@echo ">> building whisper.cpp (static, Metal)"
	cmake -S $(WHISPER_DIR) -B $(WHISPER_BUILD) \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=OFF \
		-DGGML_METAL=ON \
		-DGGML_METAL_EMBED_LIBRARY=ON \
		-DWHISPER_BUILD_EXAMPLES=OFF \
		-DWHISPER_BUILD_TESTS=OFF \
		-DWHISPER_BUILD_SERVER=OFF
	cmake --build $(WHISPER_BUILD) --config Release -j

## model: download the speech model into the app's models dir (skipped if present).
model:
	@if [ -f "$(MODEL_PATH)" ]; then \
		echo ">> model already present: $(MODEL_PATH)"; \
	else \
		mkdir -p "$(MODELS_DIR)"; \
		echo ">> downloading $(MODEL_NAME) (~1.5GB)"; \
		curl -L --fail -o "$(MODEL_PATH)" "$(MODEL_URL)"; \
	fi

## build: compile the Go binary. Builds whisper.cpp first (cgo links it) and
## generates the embedded menu-bar icons.
build: $(WHISPER_LIB) internal/app/tray-idle.png internal/app/tray-rec.png
	go build -o $(DIST)/$(EXECUTABLE) ./cmd/voysnap

# Menu-bar icons embedded into the binary via go:embed.
internal/app/tray-idle.png: build/icongen/main.go
	go run ./build/icongen bar-idle $@
internal/app/tray-rec.png: build/icongen/main.go
	go run ./build/icongen bar-rec $@

# App icon (.icns) built from the rendered master PNG at all required sizes.
build/icon.icns: build/icongen/main.go
	go run ./build/icongen app build/icon-master.png
	rm -rf build/icon.iconset && mkdir -p build/icon.iconset
	sips -z 16 16   build/icon-master.png --out build/icon.iconset/icon_16x16.png      >/dev/null
	sips -z 32 32   build/icon-master.png --out build/icon.iconset/icon_16x16@2x.png   >/dev/null
	sips -z 32 32   build/icon-master.png --out build/icon.iconset/icon_32x32.png      >/dev/null
	sips -z 64 64   build/icon-master.png --out build/icon.iconset/icon_32x32@2x.png   >/dev/null
	sips -z 128 128 build/icon-master.png --out build/icon.iconset/icon_128x128.png    >/dev/null
	sips -z 256 256 build/icon-master.png --out build/icon.iconset/icon_128x128@2x.png >/dev/null
	sips -z 256 256 build/icon-master.png --out build/icon.iconset/icon_256x256.png    >/dev/null
	sips -z 512 512 build/icon-master.png --out build/icon.iconset/icon_256x256@2x.png >/dev/null
	sips -z 512 512 build/icon-master.png --out build/icon.iconset/icon_512x512.png    >/dev/null
	cp build/icon-master.png build/icon.iconset/icon_512x512@2x.png
	iconutil -c icns build/icon.iconset -o build/icon.icns
	rm -rf build/icon.iconset build/icon-master.png
	@echo ">> built build/icon.icns"

## icon: force-regenerate all icons.
icon:
	rm -f build/icon.icns internal/app/tray-idle.png internal/app/tray-rec.png
	$(MAKE) build/icon.icns internal/app/tray-idle.png internal/app/tray-rec.png

## bundle: assemble dist/Voysnap.app around the built binary.
bundle: build build/icon.icns
	@rm -rf $(APP_DIR)
	@mkdir -p $(MACOS_DIR) $(RES_DIR)
	cp $(DIST)/$(EXECUTABLE) $(MACOS_DIR)/$(EXECUTABLE)
	cp build/Info.plist $(CONTENTS)/Info.plist
	cp build/icon.icns $(RES_DIR)/icon.icns
	@echo ">> bundled $(APP_DIR)"

# Stable self-signed identity that keeps TCC permission grants across rebuilds.
SIGN_IDENTITY := Voysnap Local Dev
SIGN_KEYCHAIN := voysnap-signing.keychain-db

## signing-identity: create the stable self-signed identity if missing (idempotent).
signing-identity:
	@security find-identity "$(SIGN_KEYCHAIN)" 2>/dev/null | grep -q "$(SIGN_IDENTITY)" \
		|| bash build/setup-signing.sh

## sign: codesign with the stable identity (or ad-hoc) so TCC grants persist.
sign: signing-identity bundle
	@if security find-identity "$(SIGN_KEYCHAIN)" 2>/dev/null | grep -q "$(SIGN_IDENTITY)"; then \
		echo ">> signing with '$(SIGN_IDENTITY)'"; \
		security unlock-keychain -p voysnap "$(SIGN_KEYCHAIN)" 2>/dev/null || true; \
		codesign --force --deep --sign "$(SIGN_IDENTITY)" --keychain "$(SIGN_KEYCHAIN)" \
			--entitlements build/entitlements.plist $(APP_DIR); \
	else \
		echo ">> no stable identity found; signing ad-hoc (run build/setup-signing.sh for persistent permissions)"; \
		codesign --force --deep --sign - --entitlements build/entitlements.plist $(APP_DIR); \
	fi
	@codesign --verify --verbose $(APP_DIR) || true

## app: full build → bundle → sign.
app: sign

## run: launch the app from dist (relaunches a fresh instance). Pulls the model first.
run: app model
	@killall $(EXECUTABLE) 2>/dev/null || true
	open $(APP_DIR)

## install: copy Voysnap.app into /Applications so it launches like any other app.
install: app model
	@killall $(EXECUTABLE) 2>/dev/null || true
	rm -rf /Applications/$(APP_NAME).app
	cp -R $(APP_DIR) /Applications/$(APP_NAME).app
	@echo ">> installed to /Applications/$(APP_NAME).app — launch it from Launchpad/Spotlight"

clean-app:
	rm -rf $(DIST)

clean: clean-app
	rm -rf $(WHISPER_BUILD)
