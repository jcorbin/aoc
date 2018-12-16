package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"os"
	"path"
	"sort"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/aoc/internal/geom"
	"github.com/jcorbin/aoc/internal/quadindex"
)

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		// for _, s := range []string{
		// } {
		// 	s = strings.Replace(s, "_", " ", -1)
		// 	io.WriteString(out, s)
		// }
		fmt.Fprintf(out, "\n\nUsage %s [options] [<inputFile>]\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	anansi.MustRun(run(os.Stdin, os.Stdout))
}

var (
	// We need to handle these signals so that we restore terminal state
	// properly (raw mode and exit the alternate screen).
	halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)

	// terminal resize signals
	resize = anansi.Notify(syscall.SIGWINCH)

	// input availability notification
	inputReady anansi.InputSignal

	// The virtual screen that will be our canvas.
	screen anansi.Screen
)

func run(in, out *os.File) error {
	var game gameUI

	if err := func() error {
		if !anansi.IsTerminal(in) {
			return game.world.load(in)
		}

		name := flag.Arg(0)
		if name == "" {
			return game.world.load(bytes.NewReader([]byte(defaultWorldData)))
		}

		f, err := os.Open(name)
		if err == nil {
			err = game.world.load(f)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}
		return err
	}(); err != nil {
		return err
	}

	worldMid := game.world.bounds.Size().Div(2)
	screenMid := screen.Bounds().Size().Div(2)
	game.focus = worldMid.Sub(screenMid)

	if !anansi.IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	term := anansi.NewTerm(in, out,
		&halt,
		&resize,
		&inputReady,
		&screen,
	)
	term.SetRaw(true)
	term.AddMode(
		ansi.ModeAlternateScreen,

		ansi.ModeMouseSgrExt,
		ansi.ModeMouseBtnEvent,
		// ansi.ModeMouseAnyEvent,
	)
	resize.Send("initialize screen size")
	return term.RunWith(&game)
}

var defaultWorldData = ""

type gameType uint8

const (
	gameRender gameType = 1 << iota
	gameCollide
	gameActor
	gameHP
	gameAP
)

type gameWorld struct {
	bounds image.Rectangle
	quadindex.Index

	round int

	t  []gameType
	p  []image.Point
	z  []int
	r  []byte
	a  []ansi.SGRAttr
	hp []int
	ap []int

	noTarget bool
	actors   []int
	goblins  map[int]struct{}
	elves    map[int]struct{}
}

func (world *gameWorld) describe(id int, buf *bytes.Buffer) {
	t := world.t[id]
	fmt.Fprintf(buf, "id:%v t:%02x", id, t)
	if t&gameRender != 0 {
		fmt.Fprintf(buf, " p:%v z:%v r:%q a:%v", world.p[id], world.z[id], world.r[id], world.a[id])
	}
	if t&gameCollide != 0 {
		buf.WriteString(" collides")
	}
	if t&gameHP != 0 {
		fmt.Fprintf(buf, " hp:%v", world.hp[id])
	}
	if t&gameAP != 0 {
		fmt.Fprintf(buf, " ap:%v", world.ap[id])
	}
	buf.WriteRune('\n')
}

func (world *gameWorld) render(g anansi.Grid, viewOffset image.Point) {
	// world bounded grid
	gz := make([]int, len(g.Rune)) // TODO re-use allocation
	for id, t := range world.t {
		if id == 0 || t&gameRender == 0 {
			continue
		}

		p := world.p[id]
		sp := p.Add(viewOffset)
		if sp.X < 0 || sp.Y < 0 {
			continue
		}

		gi, ok := g.CellOffset(ansi.PtFromImage(sp))
		if !ok {
			continue
		}

		if z := world.z[id]; gz[gi] < z {
			gz[gi] = z
			g.Rune[gi] = rune(world.r[id])
			g.Attr[gi] = mergeBGColors(world.a[id], g.Attr[gi])
		} else {
			g.Attr[gi] = mergeBGColors(g.Attr[gi], world.a[id])
		}
	}

	// sidebar annotations
	offs := make([]int, g.Bounds().Dy())
	world.sortIDs(world.actors)
	var buf bytes.Buffer
	for _, id := range world.actors {
		p := world.p[id]
		sp := p.Add(viewOffset)
		if sp.X < 0 || sp.Y < 0 {
			continue
		}
		gp := ansi.PtFromImage(sp)
		if gp.In(g.Bounds()) {
			off := offs[gp.Y-1]
			if off == 0 {
				off = world.bounds.Dx() + 3 // TODO why 3?
			}
			gp.X = off
			t := world.t[id]
			buf.Reset()
			if t&gameRender != 0 {
				fmt.Fprintf(&buf, "%s", string(world.r[id]))
			}
			if t&gameHP != 0 {
				fmt.Fprintf(&buf, "(%d)", world.hp[id])
			}
			// fmt.Fprintf(&buf, "#%d", id)
			gp = writeIntoGrid(g.SubAt(gp), buf.Bytes())
			offs[gp.Y-1] = gp.X + 1
		}
	}
}

func mergeBGColors(a, b ansi.SGRAttr) ansi.SGRAttr {
	if _, hasBG := a.BG(); !hasBG {
		if c, def := b.BG(); def {
			a |= c.BG()
		}
	}
	return a
}

func (world *gameWorld) load(r io.Reader) error {
	if len(world.t) > 0 {
		panic("reload of world not supported")
	}
	world.createEntity(0)

	// scan entities from input
	if nom, ok := r.(interface{ Name() string }); ok {
		log.Printf("read input from %s", nom.Name())
	}
	sc := bufio.NewScanner(r)
	var p image.Point
	for sc.Scan() {
		line := sc.Text()
		p.X = 0
		for i := 0; i < len(line); i++ {
			world.loadCell(p, line[i])
			p.X++
		}
		p.Y++
	}

	world.bounds = world.computeBounds()

	world.actors = make([]int, 0, len(world.t))
	for id, t := range world.t {
		if t&gameActor != 0 {
			world.actors = append(world.actors, id)
		}
	}
	world.goblins = make(map[int]struct{}, len(world.actors))
	world.elves = make(map[int]struct{}, len(world.actors))
	for _, id := range world.actors {
		switch world.r[id] {
		case 'G':
			world.goblins[id] = struct{}{}
		case 'E':
			world.elves[id] = struct{}{}
		}
	}

	return sc.Err()
}

func (world *gameWorld) loadCell(p image.Point, c byte) {
	switch c {
	case '#':
		world.createWall(p)
	case '.':
		world.createFloor(p)
	case 'G':
		world.createActor(p, 'G', ansi.RGB(128, 32, 16), 200, 3)
		world.createFloor(p)
	case 'E':
		world.createActor(p, 'E', ansi.RGB(16, 128, 32), 200, 3)
		world.createFloor(p)
	}
}

func (world *gameWorld) createWall(p image.Point) int {
	id := world.createEntity(gameRender | gameCollide)
	world.p[id] = p
	world.z[id] = 2
	world.r[id] = '#'
	world.a[id] = ansi.RGB(128, 128, 128).FG() | ansi.RGB(32, 32, 32).BG()
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createFloor(p image.Point) int {
	id := world.createEntity(gameRender)
	world.p[id] = p
	world.z[id] = 1
	world.r[id] = '.'
	world.a[id] = ansi.RGB(32, 32, 32).FG() | ansi.RGB(16, 16, 16).BG()
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createActor(p image.Point, r byte, c ansi.SGRColor, hp, ap int) int {
	id := world.createEntity(gameRender | gameCollide | gameActor | gameHP | gameAP)
	world.p[id] = p
	world.z[id] = 10
	world.r[id] = r
	world.a[id] = c.FG()
	world.hp[id] = hp
	world.ap[id] = ap
	world.Index.Update(id, p)
	return id
}

func (world *gameWorld) createEntity(t gameType) int {
	// TODO reuse above
	id := len(world.t)
	world.t = append(world.t, t)
	world.p = append(world.p, image.ZP)
	world.z = append(world.z, 0)
	world.r = append(world.r, 0)
	world.a = append(world.a, 0)
	world.hp = append(world.hp, 0)
	world.ap = append(world.ap, 0)
	return id
}

func (world *gameWorld) pruneActors() {
	i := 0
	for j := 0; j < len(world.actors); j++ {
		if world.actors[j] != 0 {
			world.actors[i] = world.actors[j]
			i++
		}
	}
	world.actors = world.actors[:i]
}

func (world *gameWorld) destroyEntity(id int) {
	for i := 0; i < len(world.actors); i++ {
		if world.actors[i] == id {
			world.actors[i] = 0
		}
	}

	delete(world.goblins, id)
	delete(world.elves, id)
	world.Index.Delete(id, world.p[id])
	world.t[id] = 0
	world.p[id] = image.ZP
	world.z[id] = 0
	world.r[id] = 0
	world.a[id] = 0
	world.hp[id] = 0
	world.ap[id] = 0
}

func (world *gameWorld) computeBounds() (bounds image.Rectangle) {
	if len(world.t) == 0 {
		return bounds
	}
	id := 1
	for ; id < len(world.t); id++ {
		if world.t[id]&gameRender != 0 {
			bounds.Min = world.p[id]
			bounds.Max = world.p[id]
			break
		}
	}
	for ; id < len(world.t); id++ {
		if world.t[id]&gameRender == 0 {
			continue
		}
		p := world.p[id]
		if bounds.Min.X > p.X {
			bounds.Min.X = p.X
		}
		if bounds.Min.Y > p.Y {
			bounds.Min.Y = p.Y
		}
		if bounds.Max.X < p.X {
			bounds.Max.X = p.X
		}
		if bounds.Max.Y < p.Y {
			bounds.Max.Y = p.Y
		}
	}
	return bounds
}

func (world *gameWorld) done() bool {
	if world.noTarget {
		return true
	}
	return false
}

func (world *gameWorld) winningTeam() string {
	if len(world.goblins) == 0 {
		return "elves"
	}
	if len(world.elves) == 0 {
		return "goblins"
	}
	return "none"
}

func (world *gameWorld) remainingHP() (hp int) {
	for _, actorID := range world.actors {
		hp += world.hp[actorID]
	}
	return hp
}

func (world *gameWorld) finish() {
	log.Printf("finished after round %v: %s won with %v remaining hp",
		world.round,
		world.winningTeam(),
		world.remainingHP(),
	)
}

func (world *gameWorld) tick() bool {
	if world.done() {
		return false
	}

	world.sortIDs(world.actors)
	for _, id := range world.actors {
		if !world.act(id) {
			if world.done() {
				world.finish()
				return false
			}
		}
	}
	world.pruneActors()
	world.round++

	return true
}

func (world *gameWorld) act(actorID int) bool {
	// already killed by prior actor in this round
	if actorID == 0 {
		return true
	}

	var enemySet map[int]struct{}
	if _, isGoblin := world.goblins[actorID]; isGoblin {
		enemySet = world.elves
	} else if _, isElf := world.elves[actorID]; isElf {
		enemySet = world.goblins
	} else {
		panic(fmt.Sprintf("neither goblin nor elf #%v", actorID))
	}

	// done if no enemies left
	if len(enemySet) == 0 {
		world.noTarget = true
		return false
	}

	attackID := world.chooseAttack(actorID, enemySet)
	if attackID == 0 {
		if !world.move(actorID, enemySet) {
			return true
		}
		attackID = world.chooseAttack(actorID, enemySet)
	}

	if attackID != 0 {
		world.attack(actorID, attackID)
	}

	return true
}

func (world *gameWorld) chooseAttack(actorID int, enemySet map[int]struct{}) int {
	// weakest adjacent enemy id
	actorP := world.p[actorID]
	attackID, attackHP := 0, 0
	for _, ep := range [4]image.Point{
		actorP.Add(image.Pt(0, -1)),
		actorP.Add(image.Pt(-1, 0)),
		actorP.Add(image.Pt(1, 0)),
		actorP.Add(image.Pt(0, 1)),
	} {
		if enemyID := world.enemyAt(ep, enemySet); enemyID != 0 {
			if attackID == 0 || world.hp[enemyID] < attackHP {
				attackID = enemyID
				attackHP = world.hp[enemyID]
			}
		}
	}
	return attackID
}

func (world *gameWorld) attack(actorID, attackID int) {
	log.Printf(
		"attack %s@%v#%v => %s@%v#%v",
		string(world.r[actorID]), world.p[actorID], actorID,
		string(world.r[attackID]), world.p[attackID], attackID,
	)
	if hp := world.hp[attackID] - world.ap[actorID]; hp < 0 {
		world.destroyEntity(attackID)
	} else {
		world.hp[attackID] = hp
	}
}

func (world *gameWorld) move(actorID int, enemySet map[int]struct{}) bool {
	// TODO re-use
	var reach reachabilityScore
	reach.Init(world)

	actorP := world.p[actorID]

	// find nearest empty cells adjacent to enemies
	reach.Update(world, actorP)
	reachableIDs := make([]int, 0, len(enemySet))
	reachableD := 0
	for enemyID := range enemySet {
		ep := world.p[enemyID]
		reachableIDs, reachableD = world.updateAdjacentReach(reach, ep, reachableIDs, reachableD)
	}
	if len(reachableIDs) == 0 {
		log.Printf("nothing reachable for %s@%v#%v", string(world.r[actorID]), world.p[actorID], actorID)
		return false
	}

	// choose the first, in reading order, most reachable enemy
	world.sortIDs(reachableIDs)
	targetID := reachableIDs[0]
	targetP := world.p[targetID]

	// move to the first, in reading order, nearest adjacent cell
	reach.Update(world, targetP)
	reachableIDs, reachableD = world.updateAdjacentReach(reach, actorP, reachableIDs[:0], 0)
	if len(reachableIDs) == 0 {
		// XXX inconceivable
		log.Printf("no nearest rearchable for %s@%v#%v targeting %s@%v#%v",
			string(world.r[actorID]), actorP, actorID,
			string(world.r[targetID]), targetP, targetID,
		)
		return false
	}
	world.sortIDs(reachableIDs)

	cellID := reachableIDs[0]
	if cellID == 0 {
		log.Printf("zero cell id! in %v", reachableIDs) // XXX inconceivable
		return false
	}

	cellP := world.p[cellID]
	log.Printf(
		"move %s@%v#%v => #%v@%v",
		string(world.r[actorID]), world.p[actorID], actorID,
		cellID, cellP,
	)
	world.p[actorID] = cellP
	world.Index.Update(actorID, cellP)

	return true
}

func (world *gameWorld) enemyAt(p image.Point, enemySet map[int]struct{}) int {
	cur := world.Index.At(p)
	for cur.Next() {
		id := cur.I()
		if _, isEnemy := enemySet[id]; isEnemy {
			return id
		}
	}
	return 0
}

func (world *gameWorld) empty(p image.Point) (id int) {
	cur := world.Index.At(p)
	for cur.Next() {
		if world.t[cur.I()]&gameCollide != 0 {
			return 0
		}
		id = cur.I()
	}
	return id
}

type reachabilityScore struct {
	geom.RCore
	sc []int
}

func (r reachabilityScore) Get(p image.Point) int {
	if i, ok := r.Index(p); ok {
		return r.sc[i]
	}
	return -1
}

type searchCore struct {
	p []image.Point // frontier pos
	d []int         // frontier distance
	// u map[int]struct{} // frontier pruning

	sort sort.Interface
}

func (srch *searchCore) init(n int) {
	if cap(srch.p) < n {
		srch.p = make([]image.Point, 0, n)
	}
	if cap(srch.d) < n {
		srch.d = make([]int, 0, n)
	}
	// if len(srch.u) == 0 {
	// 	srch.u = make(map[int]struct{}, n)
	// } else {
	// 	for id := range srch.u {
	// 		delete(srch.u, id)
	// 	}
	// }
}

func (srch *searchCore) Len() int { return len(srch.p) }

func (srch *searchCore) push(p image.Point, d, id int) {
	srch.p = append(srch.p, p)
	srch.d = append(srch.d, d)
	// if id != 0 {
	// 	srch.u[id] = struct{}{}
	// }
	if srch.sort != nil && len(srch.p) > 1 {
		sort.Sort(srch.sort)
	}
}

func (srch *searchCore) pop() (image.Point, int) {
	i := len(srch.p) - 1
	pos, d := srch.p[i], srch.d[i]
	srch.p, srch.d = srch.p[:i], srch.d[:i]
	return pos, d
}

func (r *reachabilityScore) Init(world *gameWorld) {
	r.Rectangle = world.bounds
	r.Origin = world.bounds.Min
	r.Stride = world.bounds.Dx()
	r.sc = make([]int, world.bounds.Dy()*r.Stride)
}

func (r *reachabilityScore) Update(world *gameWorld, p image.Point) {
	for i := range r.sc {
		r.sc[i] = -1
	}

	// TODO factor out reachabilitySearch

	var srch searchCore // TODO re-use
	srch.init(len(world.p))
	srch.push(p, 0, world.empty(p))
	ri, _ := r.Index(p)
	r.sc[ri] = 0

	// sort reverse by distance, so we expand (a) closest point first
	srch.sort = sortPointRevBy{sortPointBy{&srch.p, &srch.d}}

	for srch.Len() > 0 {
		pos, d := srch.pop()

		// skip if it's not better than a prior score
		ri, _ := r.Index(pos)
		if sc := r.sc[ri]; sc >= 0 && d > sc {
			continue
		}

		// ... expand to any adjacent empty cells that are better than prior
		pts, cellIDs := world.adjacentCells(pos)
		for i, cellID := range cellIDs {
			if cellID != 0 {
				qp := pts[i]
				ri, _ := r.Index(qp)
				if sc := r.sc[ri]; sc < 0 || sc > d+1 {
					srch.push(qp, d+1, cellID)
					r.sc[ri] = d + 1
				}
			}
		}
	}
}

func (world *gameWorld) updateAdjacentReach(
	reach reachabilityScore,
	from image.Point,
	ids []int, best int,
) ([]int, int) {
	pts, cellIDs := world.adjacentCells(from)
	for i, cellID := range cellIDs {
		if cellID != 0 {
			d := reach.Get(pts[i])
			if d < 0 {
				continue
			}
			if best > 0 && best > d {
				ids = ids[:0]
			}
			ids = append(ids, cellID)
			best = d
		}
	}
	return ids, best
}

func (world *gameWorld) adjacentCells(p image.Point) (pts [4]image.Point, cellIDs [4]int) {
	pts[0] = p.Add(image.Pt(0, -1))
	pts[1] = p.Add(image.Pt(-1, 0))
	pts[2] = p.Add(image.Pt(1, 0))
	pts[3] = p.Add(image.Pt(0, 1))
	cellIDs[0] = world.empty(pts[0])
	cellIDs[1] = world.empty(pts[1])
	cellIDs[2] = world.empty(pts[2])
	cellIDs[3] = world.empty(pts[3])
	return pts, cellIDs
}

func manhattanDistance(a, b image.Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (world *gameWorld) sortIDs(ids []int) {
	sort.Slice(ids, func(i, j int) bool {
		idi, idj := ids[i], ids[j]
		pi, pj := world.p[idi], world.p[idj]
		if pi.Y < pj.Y {
			return true
		}
		return pi.Y == pj.Y && pi.X < pj.X
	})
}

func (world *gameWorld) handleInput(term *anansi.Term, game *gameUI) error {
	// switch e {
	// TODO
	// }
	return nil
}

type gameUI struct {
	banner []byte

	timer    *time.Timer
	last     time.Time
	ticking  bool
	playing  bool
	playRate int // tick-per-second

	focus      image.Point
	viewOffset image.Point

	mess     []byte
	messSize image.Point

	world gameWorld
}

func (game *gameUI) Run(term *anansi.Term) error {
	game.timer = time.NewTimer(100 * time.Second)
	game.stopTimer()
	return term.Loop(game)
}

func (game *gameUI) setTimer(d time.Duration) {
	if game.timer == nil {
		game.timer = time.NewTimer(d)
	} else {
		game.timer.Reset(d)
	}
}

func (game *gameUI) stopTimer() {
	game.timer.Stop()
	select {
	case <-game.timer.C:
	default:
	}
}

func (game *gameUI) Update(term *anansi.Term) (redraw bool, _ error) {
	select {
	case sig := <-halt.C:
		return false, anansi.SigErr(sig)

	case <-resize.C:
		if err := screen.SizeToTerm(term); err != nil {
			return false, err
		}
		redraw = true

	case <-inputReady.C:
		_, err := term.ReadAny()
		herr := game.handleInput(term)
		if err == nil {
			err = herr
		}
		if err != nil {
			return false, err
		}

	case now := <-game.timer.C:
		game.advance(now)
		redraw = true
	}
	return redraw, nil
}

func (game *gameUI) advance(now time.Time) {
	// no updates while displaying a message
	if !game.ticking {
		game.last = now
		return
	}

	// single-step
	if !game.playing {
		game.world.tick()
		game.last = now
		return
	}

	// advance playback
	if ticks := int(math.Round(float64(now.Sub(game.last)) / float64(time.Second) * float64(game.playRate))); ticks > 0 {
		const maxTicks = 100000
		if ticks > maxTicks {
			ticks = maxTicks
		}
		for i := 0; i < ticks; i++ {
			if !game.world.tick() {
				game.playing = false
				break
			}
		}
		game.last = now
	}

	game.ticking = true
	game.setTimer(10 * time.Millisecond) // TODO compute next time when ticks > 0; avoid spurious wakeup
}

func (game *gameUI) setBanner(mess string, args ...interface{}) {
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	game.banner = []byte(mess)
}

func (game *gameUI) handleInput(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		for _, handlerFn := range []func(e ansi.Escape, a []byte) (bool, error){
			game.handleLowInput,
			game.handleMessInput,
			game.handleWorldInput,
		} {
			if handled, err := handlerFn(e, a); err != nil {
				return err
			} else if handled {
				break
			}
		}
	}
	return nil
}

func (game *gameUI) handleLowInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	case 0x03: // stop on Ctrl-C
		return true, fmt.Errorf("read %v", e)

	case 0x0c: // clear screen on Ctrl-L
		screen.Clear()           // clear virtual contents
		screen.To(ansi.Pt(1, 1)) // cursor back to top
		screen.Invalidate()      // force full redraw
		game.setTimer(5 * time.Millisecond)
		return true, nil

	}
	return false, nil
}

func (game *gameUI) handleMessInput(e ansi.Escape, a []byte) (bool, error) {
	// no message, ignore
	if game.mess == nil {
		return false, nil
	}

	switch e {

	// <Esc> to dismiss message
	case ansi.Escape('\x1b'):
		game.setMess(nil)
		return true, nil

	// eat any other input when a message is shown
	default:
		return true, nil
	}
}

func (game *gameUI) handleWorldInput(e ansi.Escape, a []byte) (bool, error) {
	switch e {

	// TODO
	// // display help
	// case ansi.Escape('?'):
	// 	return true, nil

	// arrow keys to move view
	case ansi.CUB, ansi.CUF, ansi.CUU, ansi.CUD:
		if d, ok := ansi.DecodeCursorCardinal(e, a); ok {
			p := game.focus.Add(d)
			if p.X < game.world.bounds.Min.X {
				p.X = game.world.bounds.Min.X
			}
			if p.Y < game.world.bounds.Min.Y {
				p.Y = game.world.bounds.Min.Y
			}
			if p.X >= game.world.bounds.Max.X {
				p.X = game.world.bounds.Max.X - 1
			}
			if p.Y >= game.world.bounds.Max.Y {
				p.Y = game.world.bounds.Max.Y - 1
			}
			if game.focus != p {
				game.focus = p
				game.setTimer(5 * time.Millisecond)
			}
		}
		return true, nil

	// mouse inspection
	case ansi.CSI('m'), ansi.CSI('M'):
		if m, sp, err := ansi.DecodeXtermExtendedMouse(e, a); err == nil {
			if m.ButtonID() == 1 && m.IsRelease() {
				var buf bytes.Buffer
				buf.Grow(1024)

				p := sp.ToImage().Sub(game.viewOffset)
				fmt.Fprintf(&buf, "Query @%v\n", p)

				n := 0
				cur := game.world.Index.At(p)
				for i := 0; cur.Next(); i++ {
					game.world.describe(cur.I(), &buf)
					n++
				}
				if n == 0 {
					fmt.Fprintf(&buf, "No Results\n")
				}
				fmt.Fprintf(&buf, "( <Esc> to close )")

				game.setMess(buf.Bytes()) // TODO w/ mouse handler rentrance
			}
		}
		return true, nil

	// step
	case ansi.Escape('.'):
		game.setTimer(5 * time.Millisecond)
		game.ticking = true
		return true, nil

	// play/pause
	case ansi.Escape(' '):
		game.playing = !game.playing
		if !game.playing {
			game.stopTimer()
			log.Printf("pause")
		} else {
			game.last = time.Now()
			if game.playRate == 0 {
				game.playRate = 1
			}
			game.ticking = true
			log.Printf("play at %v ticks/s", game.playRate)
		}
		game.setTimer(5 * time.Millisecond)
		return true, nil

	// speed control
	case ansi.Escape('+'):
		game.playRate *= 2
		log.Printf("speed up to %v ticks/s", game.playRate)
		game.setTimer(5 * time.Millisecond)
		return true, nil
	case ansi.Escape('-'):
		rate := game.playRate / 2
		if rate <= 0 {
			rate = 1
		}
		if game.playRate != rate {
			game.playRate = rate
			log.Printf("slow down to %v ticks/s", game.playRate)
		}
		game.setTimer(5 * time.Millisecond)
		return true, nil

	}

	return false, nil
}

func (game *gameUI) WriteTo(w io.Writer) (n int64, err error) {
	screen.Clear()
	game.world.render(screen.Grid, game.viewOffset)
	overlayLogs()
	game.overlayBanner()
	game.overlayMess()
	return screen.WriteTo(w)
}

func (game *gameUI) setMess(mess []byte) {
	game.mess = mess
	if mess == nil {
		game.messSize = image.ZP
	} else {
		game.messSize = measureTextBox(mess).Size()
	}
	game.setTimer(5 * time.Millisecond)
}

func (game *gameUI) overlayBanner() {
	at := screen.Grid.Rect.Min
	bannerWidth := measureTextBox(game.banner).Dx()
	screenWidth := screen.Bounds().Dx()
	at.X += screenWidth/2 - bannerWidth/2
	writeIntoGrid(screen.Grid.SubAt(at), game.banner)
}

func (game *gameUI) overlayMess() {
	if game.mess == nil || game.messSize == image.ZP {
		return
	}
	screenSize := screen.Bounds().Size()
	screenMid := screenSize.Div(2)
	messMid := game.messSize.Div(2)
	offset := screenMid.Sub(messMid)
	writeIntoGrid(screen.Grid.SubAt(screen.Grid.Rect.Min.Add(offset)), game.mess)
}

func writeIntoGrid(g anansi.Grid, b []byte) ansi.Point {
	var cur anansi.CursorState
	cur.Point = g.Rect.Min
	for len(b) > 0 {
		e, a, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {
		case ansi.Escape('\n'):
			cur.Y++
			cur.X = g.Rect.Min.X

		case ansi.CSI('m'):
			if attr, _, err := ansi.DecodeSGR(a); err == nil {
				cur.MergeSGR(attr)
			}

		default:
			// write runes into grid, with cursor style, ignoring any other
			// escapes; treating `_` as transparent
			if !e.IsEscape() {
				if i, ok := g.CellOffset(cur.Point); ok {
					if e != ansi.Escape('_') {
						g.Rune[i] = rune(e)
						g.Attr[i] = cur.Attr
					}
				}
				cur.X++
			}
		}
	}
	return cur.Point
}

func measureTextBox(b []byte) (box ansi.Rectangle) {
	box.Min = ansi.Pt(1, 1)
	box.Max = ansi.Pt(1, 1)
	pt := box.Min
	for len(b) > 0 {
		e, _ /*a*/, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e == 0 {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			e = ansi.Escape(r)
		}
		switch e {

		case ansi.Escape('\n'):
			pt.Y++
			pt.X = 1

			// TODO would be nice to borrow cursor movement processing from anansi.Screen et al

		default:
			// ignore escapes, advance on runes
			if !e.IsEscape() {
				pt.X++
			}

		}
		if box.Max.X < pt.X {
			box.Max.X = pt.X
		}
		if box.Max.Y < pt.Y {
			box.Max.Y = pt.Y
		}
	}
	return box
}

// in-memory log buffer
var logs bytes.Buffer

func init() {
	f, err := os.Create("game.log")
	if err != nil {
		log.Fatalf("failed to create game.log: %v", err)
	}
	log.SetOutput(io.MultiWriter(
		&logs,
		f,
	))
}

func overlayLogs() {
	n := bytes.Count(logs.Bytes(), []byte{'\n'})
	lb := logs.Bytes()
	for n > 10 {
		off := bytes.IndexByte(lb, '\n')
		if off < 0 {
			break
		}
		lb = lb[off+1:]
		logs.Next(off + 1)
		n--
	}

	screen.To(ansi.Pt(1, screen.Bounds().Dy()-n+1))

	lb = logs.Bytes()
	for {
		off := bytes.IndexByte(lb, '\n')
		if off < 0 {
			screen.Write(lb)
			break
		}
		screen.Write(lb[:off])
		screen.WriteString("\r\n")
		lb = lb[off+1:]
	}
}

type sortBy struct{ a, b []int }

func (s sortBy) Len() int               { return len(s.a) }
func (s sortBy) Less(i int, j int) bool { return s.b[i] < s.b[j] }
func (s sortBy) Swap(i int, j int) {
	s.a[i], s.a[j] = s.a[j], s.a[i]
	s.b[i], s.b[j] = s.b[j], s.b[i]
}

type sortPointBy struct {
	a *[]image.Point
	b *[]int
}

func (s sortPointBy) Len() int               { return len(*(s.a)) }
func (s sortPointBy) Less(i int, j int) bool { return (*s.b)[i] < (*s.b)[j] }
func (s sortPointBy) Swap(i int, j int) {
	(*s.a)[i], (*s.a)[j] = (*s.a)[j], (*s.a)[i]
	(*s.b)[i], (*s.b)[j] = (*s.b)[j], (*s.b)[i]
}

type sortPointRevBy struct{ sortPointBy }

func (s sortPointRevBy) Less(i int, j int) bool { return (*s.b)[i] > (*s.b)[j] }
