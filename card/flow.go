package card

import (
	"fmt"
	"strings"
)

// run is one dot's trip: it leaves at start (seconds into the loop) and is in
// transit for dur seconds.
type run struct {
	start   float64
	dur     float64
	reverse bool
	bounce  bool // out and back
}

// flowRuns expands a flow into its individual dots. It is the single source of
// truth for a flow's timing: rendering and the step scheduler both use it, so
// the derived loop length can never disagree with what is drawn.
func flowRuns(f *Flow, travel float64) []run {
	var runs []run
	for k := 0; k < f.Count; k++ {
		s := f.at + float64(k)*f.Gap
		switch f.Mode {
		case FlowReverse:
			runs = append(runs, run{start: s, dur: travel, reverse: true})
		case FlowAlternate:
			runs = append(runs, run{start: s, dur: 2*travel + f.Pause, bounce: true})
		case FlowBoth:
			// Independent, deliberately out of phase so it reads as full
			// duplex rather than a synchronized request/response.
			runs = append(runs,
				run{start: s, dur: travel},
				run{start: s + travel*0.5, dur: travel, reverse: true})
		default: // forward
			runs = append(runs, run{start: s, dur: travel})
		}
	}
	return runs
}

// flowSpan is how long the flow takes from its own start to its last dot
// arriving, in seconds.
func flowSpan(f *Flow, travel float64) float64 {
	var end float64
	for _, r := range flowRuns(f, travel) {
		end = max(end, r.start+r.dur-f.at)
	}
	return end
}

// flowDots renders the SMIL circles for one flow on an edge. Every dot's
// animation has period `loop`, so all edges repeat together and a GIF/video
// loop is seamless. Dots are hidden (opacity 0) except while travelling, which
// also keeps them off the top-left corner in renderers that ignore SMIL.
func flowDots(pathID string, e Edge, f *Flow, q quad, loop float64) (string, error) {
	var b strings.Builder
	for _, r := range flowRuns(f, q.length()/f.Speed) {
		if r.start < 0 || r.start+r.dur > loop+1e-9 {
			return "", fmt.Errorf("edge %q: flow needs %.2fs starting at %.2fs but the loop is %.2fs; raise loop or leave it unset to derive it",
				e.ID, r.dur, r.start, loop)
		}
		b.WriteString(dotSVG(pathID, f, loop, r))
	}
	return b.String(), nil
}

// dotSVG emits one dot: a circle with a halo, moved by animateMotion along the
// edge path and shown only while in transit.
func dotSVG(pathID string, f *Flow, loop float64, r run) string {
	end := r.start + r.dur
	var keyTimes, keyPoints []float64
	switch {
	case r.bounce:
		travel := (r.dur - f.Pause) / 2
		out, back := r.start+travel, r.start+travel+f.Pause
		keyTimes = []float64{0, r.start, out, back, end, loop}
		keyPoints = []float64{0, 0, 1, 1, 0, 0}
	case r.reverse:
		keyTimes = []float64{0, r.start, end, loop}
		keyPoints = []float64{1, 1, 0, 0}
	default:
		keyTimes = []float64{0, r.start, end, loop}
		keyPoints = []float64{0, 0, 1, 1}
	}
	for i := range keyTimes {
		keyTimes[i] /= loop
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<circle r="%s" fill="%s" stroke="%s" stroke-opacity="0.35" stroke-width="%s" opacity="0">`,
		num(f.Size), esc(f.Color), esc(f.Color), num(f.Size*0.9))
	fmt.Fprintf(&b, `<animateMotion dur="%ss" repeatCount="indefinite" calcMode="linear" keyPoints="%s" keyTimes="%s"><mpath href="#%s"/></animateMotion>`,
		num(loop), joinNums(keyPoints), joinNums(keyTimes), pathID)
	fmt.Fprintf(&b, `<animate attributeName="opacity" dur="%ss" repeatCount="indefinite" calcMode="discrete" values="0;1;0" keyTimes="%s"/>`,
		num(loop), joinNums([]float64{0, r.start / loop, end / loop}))
	b.WriteString(`</circle>`)
	return b.String()
}
