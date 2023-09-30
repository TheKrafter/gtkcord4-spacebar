package directbutton

import (
	"context"
	"log"
	"sort"

	"github.com/thekrafter/arikawa-spacebar/v3/discord"
	"github.com/thekrafter/arikawa-spacebar/v3/gateway"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotkit/gtkutil"
	"github.com/diamondburned/gotkit/gtkutil/cssutil"
	"github.com/thekrafter/gtkcord4-spacebar/internal/gtkcord"
	"github.com/diamondburned/ningen/v3/states/read"
)

// Opener is the interface that a controller must implement to open a channel.
// It is solely activated by user interaction.
type Opener interface {
	// OpenDMs opens the DMs view.
	OpenDMs()
	// OpenChannel opens the channel with the given ID.
	OpenChannel(discord.ChannelID)
}

type View struct {
	*gtk.Box
	DM *Button

	mentioned struct {
		IDs     []discord.ChannelID
		Buttons map[discord.ChannelID]*ChannelButton
	}

	ctx    context.Context
	opener Opener
}

var viewCSS = cssutil.Applier("dmbutton-view", `
`)

func NewView(ctx context.Context, opener Opener) *View {
	v := View{
		Box: gtk.NewBox(gtk.OrientationVertical, 0),
		DM:  NewButton(ctx, opener.OpenDMs),

		ctx:    ctx,
		opener: opener,
	}

	v.mentioned.IDs = make([]discord.ChannelID, 0, 4)
	v.mentioned.Buttons = make(map[discord.ChannelID]*ChannelButton, 4)

	v.Append(v.DM)
	viewCSS(v)

	vis := gtkutil.WithVisibility(ctx, v)

	state := gtkcord.FromContext(ctx)
	state.BindHandler(vis, func(ev gateway.Event) {
		switch ev := ev.(type) {
		case *read.UpdateEvent:
			if !ev.GuildID.IsValid() {
				v.Invalidate()
			}
		case *gateway.MessageCreateEvent:
			if !ev.GuildID.IsValid() {
				v.Invalidate()
			}
		}
	},
		(*read.UpdateEvent)(nil),
		(*gateway.MessageCreateEvent)(nil),
	)

	return &v
}

type channelUnreadStatus struct {
	*discord.Channel
	UnreadCount int
}

func (v *View) Invalidate() {
	state := gtkcord.FromContext(v.ctx)

	// This is slow, but whatever.
	dms, err := state.PrivateChannels()
	if err != nil {
		log.Println("dmbutton.View: failed to get private channels:", err)

		// Clear all DMs.
		v.update(nil)
		return
	}

	var unreads map[discord.ChannelID]channelUnreadStatus
	for i, dm := range dms {
		count := state.ChannelCountUnreads(dm.ID)
		if count == 0 {
			continue
		}

		if unreads == nil {
			unreads = make(map[discord.ChannelID]channelUnreadStatus, 4)
		}

		unreads[dm.ID] = channelUnreadStatus{
			Channel:     &dms[i],
			UnreadCount: count,
		}
	}

	v.update(unreads)
}

func (v *View) update(unreads map[discord.ChannelID]channelUnreadStatus) {
	for _, unread := range unreads {
		button, ok := v.mentioned.Buttons[unread.Channel.ID]
		if !ok {
			button = NewChannelButton(v.ctx, unread.Channel.ID, v.opener)
			v.mentioned.Buttons[unread.Channel.ID] = button
		}

		button.Update(unread.Channel)
		button.InvalidateUnread()
	}

	// Purge all buttons off the widget.
	for _, button := range v.mentioned.Buttons {
		v.Remove(button)
	}

	// Delete unused buttons.
	for id := range v.mentioned.Buttons {
		if _, ok := unreads[id]; !ok {
			delete(v.mentioned.Buttons, id)
		}
	}

	// Recreate the IDs slice.
	v.mentioned.IDs = v.mentioned.IDs[:0]
	for id := range unreads {
		v.mentioned.IDs = append(v.mentioned.IDs, id)
	}

	// Sort the IDs slice. We'll sort it according to the time that the last
	// message was sent: the most recent message will be at the top.
	sort.Slice(v.mentioned.IDs, func(i, j int) bool {
		mi := unreads[v.mentioned.IDs[i]]
		mj := unreads[v.mentioned.IDs[j]]
		return mi.LastMessageID > mj.LastMessageID
	})

	// Append the buttons back to the widget.
	for _, id := range v.mentioned.IDs {
		v.Append(v.mentioned.Buttons[id])
	}
}

func (v *View) Unselect() {
	v.DM.Pill.State = 0
	v.DM.Pill.Invalidate()
}
