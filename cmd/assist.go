package cmd

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"
	"time"

	xdraw "golang.org/x/image/draw"

	"github.com/ideamans/apple-app-store-connect-cli/internal/api"
)

// --- asset delivery polling ----------------------------------------------------
//
// Committing an asset upload (PATCH uploaded=true) returning 2xx does NOT mean
// the asset was accepted: Apple validates asynchronously and the result only
// shows up in assetDeliveryState (e.g. FAILED / IMAGE_INCORRECT_DIMENSIONS).

// Poll pacing, variables so tests can tighten them.
var (
	assetPollInterval = 3 * time.Second
	assetPollTimeout  = 90 * time.Second
)

// waitAssetDelivery polls resourcePath (e.g. "/v1/appScreenshots/<id>") until
// the delivery state is terminal. FAILED returns an error carrying the API's
// error codes; COMPLETE (or a resource without a delivery state) returns nil.
// Problems with the polling itself only warn — the upload already succeeded.
func waitAssetDelivery(ctx context.Context, c *api.Client, resourcePath string) error {
	if c.DryRun {
		return nil
	}
	deadline := time.Now().Add(assetPollTimeout)
	announced := false
	for {
		res, _, err := c.Get(ctx, resourcePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not check delivery state of %s: %v\n", resourcePath, err)
			return nil
		}
		state, detail := assetDeliveryInfo(res)
		switch state {
		case "", "COMPLETE":
			return nil
		case "FAILED":
			msg := "asset validation failed"
			if detail != "" {
				msg += ": " + detail
			}
			return fmt.Errorf("%s (the 2xx upload response only reserves the asset; validation is asynchronous)", msg)
		}
		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "warning: %s still %s after %s; check the final state later with the matching list command\n", resourcePath, state, assetPollTimeout)
			return nil
		}
		if !announced {
			fmt.Printf("  waiting for Apple-side validation (%s)...\n", state)
			announced = true
		}
		time.Sleep(assetPollInterval)
	}
}

// assetDeliveryInfo extracts the delivery state and a joined error summary from
// assetDeliveryState (or videoDeliveryState for app previews).
func assetDeliveryInfo(r *api.Resource) (state, detail string) {
	for _, key := range []string{"assetDeliveryState", "videoDeliveryState"} {
		ds, ok := r.Attributes[key].(map[string]any)
		if !ok {
			continue
		}
		if s, ok := ds["state"].(string); ok && s != "" {
			state = s
		}
		if errs, ok := ds["errors"].([]any); ok {
			var parts []string
			for _, e := range errs {
				if em, ok := e.(map[string]any); ok {
					code, _ := em["code"].(string)
					desc, _ := em["description"].(string)
					switch {
					case code != "" && desc != "":
						parts = append(parts, code+" ("+desc+")")
					case code != "":
						parts = append(parts, code)
					case desc != "":
						parts = append(parts, desc)
					}
				}
			}
			detail = strings.Join(parts, "; ")
		}
		if state != "" {
			return state, detail
		}
	}
	return state, detail
}

// --- image dimension assistance --------------------------------------------------
//
// The review-screenshot validator for in-app purchases and subscriptions only
// accepts legacy device sizes; App Store screenshots accept current sizes. See
// feedback from real submissions: modern sizes (1290×2796 etc.) all fail with
// IMAGE_INCORRECT_DIMENSIONS on inAppPurchaseAppStoreReviewScreenshots.

// reviewShotSizes are the known-accepted pixel sizes for IAP/subscription
// review screenshots (legacy 5.5" iPhone and 12.9" iPad, both orientations).
var reviewShotSizes = [][2]int{
	{1242, 2208}, {2208, 1242},
	{2048, 2732}, {2732, 2048},
}

func reviewShotSizeList() string {
	parts := make([]string, len(reviewShotSizes))
	for i, s := range reviewShotSizes {
		parts[i] = fmt.Sprintf("%d×%d", s[0], s[1])
	}
	return strings.Join(parts, ", ")
}

// imageDims returns the pixel dimensions of PNG/JPEG data.
func imageDims(data []byte) (w, h int, err error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("decode image: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

// prepareReviewScreenshot validates (and with autoFit, converts) an IAP or
// subscription review screenshot. Without autoFit an unsupported size errors
// out before any bytes are uploaded; with autoFit the image is scaled to fit
// 1242×2208 (portrait) or 2208×1242 (landscape) preserving aspect ratio, and
// padded with white. The returned file name gets a .png extension when the
// data was re-encoded.
func prepareReviewScreenshot(path string, autoFit bool) (data []byte, fileName string, err error) {
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	fileName = baseName(path)
	w, h, err := imageDims(data)
	if err != nil {
		return nil, "", err
	}
	for _, s := range reviewShotSizes {
		if w == s[0] && h == s[1] {
			return data, fileName, nil
		}
	}
	if !autoFit {
		return nil, "", fmt.Errorf(
			"%s is %d×%d, which the review-screenshot validator rejects (App Store screenshot sizes do not apply here).\nAccepted sizes: %s.\nPass --auto-fit to scale and pad it to %s automatically",
			fileName, w, h, reviewShotSizeList(), fitTargetLabel(w, h))
	}
	tw, th := 1242, 2208
	if w > h {
		tw, th = 2208, 1242
	}
	fitted, err := fitImage(data, tw, th)
	if err != nil {
		return nil, "", err
	}
	fmt.Printf("auto-fit: %s %d×%d -> %d×%d (aspect kept, white padding)\n", fileName, w, h, tw, th)
	if i := strings.LastIndex(fileName, "."); i > 0 {
		fileName = fileName[:i]
	}
	return fitted, fileName + ".png", nil
}

func fitTargetLabel(w, h int) string {
	if w > h {
		return "2208×1242"
	}
	return "1242×2208"
}

// fitImage scales src to fit within tw×th preserving aspect ratio, centers it
// on a white canvas of exactly tw×th, and re-encodes as PNG.
func fitImage(data []byte, tw, th int) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	sb := src.Bounds()
	scale := min(float64(tw)/float64(sb.Dx()), float64(th)/float64(sb.Dy()))
	sw := int(float64(sb.Dx())*scale + 0.5)
	sh := int(float64(sb.Dy())*scale + 0.5)
	canvas := image.NewRGBA(image.Rect(0, 0, tw, th))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	dst := image.Rect((tw-sw)/2, (th-sh)/2, (tw-sw)/2+sw, (th-sh)/2+sh)
	xdraw.CatmullRom.Scale(canvas, dst, src, sb, xdraw.Over, nil)
	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func baseName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
