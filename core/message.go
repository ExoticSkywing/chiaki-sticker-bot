package core

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/star-39/moe-sticker-bot/pkg/msbimport"
	tele "gopkg.in/telebot.v3"
)

func sendStartMessage(c tele.Context) error {
	message := `
Hi! I'm <b>Chiaki Sticker Bot</b>! Please:
• Use <b>/import</b> or send <b>LINE/Kakao sticker share link</b> to import or download.
• Send <b>Telegram sticker/link/GIF</b> to download.
• Send <b>keywords</b> to search sticker sets.
• Use <b>/create</b> or <b>/manage</b> to create or manage sticker sets and CustomEmoji.
• Check all available commands: <b>/command_list</b>.

你好！欢迎使用 <b>Chiaki Sticker Bot</b>！请：
• 使用 <b>/import</b> 或发送 <b>LINE/Kakao 贴图包分享链接</b>来导入或下载。
• 发送 <b>Telegram 贴图／链接／GIF</b> 来下载。
• 发送 <b>/search</b> 来搜索贴图包。
• 使用 <b>/create</b> 或 <b>/manage</b> 来创建或管理贴图包和表情贴。
• 发送 <b>/command_list</b> 查看所有可用指令。
`
	return c.Send(message, tele.ModeHTML, tele.NoPreview)
}

func sendCommandList(c tele.Context) error {
	message := `
<b>/import</b>  <b>/search</b> LINE/Kakao stickers.<code>
导入或搜索 LINE/Kakao 贴图包.</code>
<b>/download</b>  <b>/create</b>  <b>/manage</b> Telegram stickers.<code>
下载、创建、管理 Telegram 贴图包.</code>
<b>/faq  /about  /changelog  /privacy</b><code>
常见问题/关于/更新记录/隐私</code>
`

	return c.Send(message, tele.ModeHTML, tele.NoPreview)
}

func sendAboutMessage(c tele.Context) {
	c.Send(fmt.Sprintf(`
<b>Chiaki Sticker Bot</b>
A Telegram sticker bot — import LINE/Kakao stickers, create and manage your own sticker sets, and download stickers with ease.

Telegram 贴图机器人，支持导入 LINE/Kakao 贴图、创建与管理贴图包，以及下载贴图。

<a href="https://github.com/akira02/chiaki-sticker-bot">GitHub: akira02/chiaki-sticker-bot</a>
Forked from the great work of <a href="https://github.com/star-39/moe-sticker-bot">star-39/moe-sticker-bot</a>.
<code>
This free(as in freedom) software is released under the GPLv3 License.
Comes with ABSOLUTELY NO WARRANTY! All rights reserved.
本 BOT 为免费提供的自由软件，你可以自由使用/分发，但不提供任何保用(warranty)！
本软件授权于通用公众授权条款(GPL)v3，保留所有权利。
</code>
Version／版本: %s
`, BOT_VERSION), tele.ModeHTML)
}

func sendFAQ(c tele.Context) {
	c.Send(fmt.Sprintf(`
<b>Please hit Star for this project on GitHub if you like this bot!
如果你喜欢这个 bot, 欢迎在 GitHub 给本项目点 Star 喔!
https://github.com/akira02/chiaki-sticker-bot</b>
------------------------------------
<b>Q: I got stucked! I can't quit from command!
我卡住了! 我没办法从指令中退出!</b>
A: Please send /quit to interrupt.
请发送 /quit 来中断。

<b>Q: Why ID has suffix: _by_%s ?
为什么 ID 的末尾有: _by_%s ?</b>
A: It's forced by Telegram, bot created sticker set must have its name in ID suffix.
因为这是 Telegram 的强制要求, 由 bot 创建的贴图 ID 末尾必须有 bot 名字。

<b>Q: Who owns the sticker sets the bot created?
    BOT 创建的贴图包由谁所有？</b>
A: It's you of course. You can manage them through /manage or Telegram's official @Stickers bot.
    当然是你。你可以通过 /manage 指令或者 Telegram 官方的 @Stickers 管理你的贴图包。
`, botName, botName), tele.ModeHTML)
}

func sendChangelog(c tele.Context) error {
	return c.Send(`
Details: 详细:
https://github.com/akira02/chiaki-sticker-bot#changelog
v2.5.0-RC1(20240528)
* Support mix-typed sticker set.
* You can add video to static set and vice versa.
* Removed WhatsApp export temporarily .
* 支持混合贴图包。
* 贴图包可以同时存入静态与动态贴图。
* 暂时移除WhatsApp导出功能。

v2.4.0-RC1-RC4(20240304)
* Support Importing LINE Emoji into CustomEmoji.
* Support creating CustomEmoji.
* Support editing sticker emoji and title.
* 支持 LINE 表情贴导入。
* 支持创建表情贴。
* 支持修改贴图 Emoji/贴图包标题。

v2.3.13-v2.3.15(20230228)
* Support region locked LINE Message sticker.
* Support TGS(Animated) sticker export.
* Fix TGS(Animated) sticker download. 
* 支持有区域锁的 LINE 消息贴图。
* 支持TGS 贴图导出。
* 修复TGS(动态)贴图下载问题.

v2.3.10(20230217)
  * Fix Kakao import fatal, support more animated Kakao.
  * 修复 Kakao 导入错误, 支持更多 Kakao 动态贴图.

v2.3.x (20230216)
  * Fix flood limit error during import.
  * Fix animated Kakao treated as static.
  * Improved static Kakao quality.
  * Support changing sticker title.
  * 修复导入贴图时flood limit错误。
  * 修复动态 Kakao 被当作静态.
  * 提升静态 Kakao 画质.
  * 支持修改贴图包标题.
  
v2.2.0 (20230131)
  * Support animated Kakao sticker.
  * 支持动态 Kakao 贴图。

v2.1.0 (20230129)
  * Support exporting sticker to WhatsApp.
  * 支持导出贴图到WhatsApp

v2.0.0 (20230105)
  * Use new WebApp from /manage command to edit sticker set with ease.
  * Send text or use /search command to search imported LINE/Kakao sticker sets by all users.
  * Auto import now happens on backgroud.
  * Downloading sticker set is now lot faster.
  * Fix many LINE import issues.
  * 通过 /manage 指令使用新的 WebApp 轻松管理贴图包.
  * 直接发送文字或使用 /search 指令来搜索所有用户导入的 LINE/Kakao 贴图包.
  * 自动导入现在会在后台处理.
  * 下载整个贴图包的速度现在会快许多.
  * 修复了许多LINE 贴图导入的问题.
	`, tele.NoPreview)
}

func sendPrivacy(c tele.Context) error {
	return c.Send(`
<b>Privacy Notice:</b>
This bot collects limited data to operate and improve the service.

<b>What we collect:</b>
• <b>Failure events</b> — if /import fails, your ID, action type, and failure reason are logged for diagnosing.
• <b>Sticker ownership</b> — if you use /import or /create successfully, your user ID is associated with the sticker set so that /manage can identify your sets.

<b>What we do NOT collect:</b>
• Message content
• Sticker files you send
• Any data from users who only browse or interact with sticker sets they did not create

All stored data is used solely to operate this bot and will never be shared with any third party.
The bot server is located in Singapore. Local laws may apply.
This bot is free and open source: https://github.com/akira02/chiaki-sticker-bot

<b>隐私声明:</b>
本 bot 会收集有限的数据以维持服务运行。

<b>收集的数据：</b>
• <b>使用记录</b> — 当你使用 /import 时，你的 ID、操作类型及失败原因会被记录，用于诊断问题。
• <b>贴图包归属</b> — 当你成功使用 /import 或 /create 时，你的用户 ID 会与所建立的贴图包关联，以供 /manage 识别你的贴图包。

<b>不收集的数据：</b>
• 消息内容
• 你发送的贴图文件
• 未使用上述指令的用户的任何数据

所有存储的数据仅用于维持本 bot 运行，不会分享给任何第三方。
本 bot 服务器位于新加坡，将适用当地法律。
本 bot 为自由开源软件：https://github.com/akira02/chiaki-sticker-bot
`, tele.ModeHTML, tele.NoPreview)
}

func sendAskEmoji(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Assign separately/分别设置", "manual")
	btnRand := selector.Data(`Batch assign as/批量设置为 "⭐"`, "random")
	selector.Inline(selector.Row(btnManu), selector.Row(btnRand))

	return c.Send(`
Telegram requires emoji to and keywords for each sticker:
• Press "Assign separately" to assign emoji and keywords one by one.
• Send an emoji to do batch assign.
Telegram 要求为每张贴图分别设置 emoji 和关键字:
• 按下"分别设置"来为每个贴图分别设置相应的 emoji 和关键字。
• 发送一个 emoji 来为全部贴图设置成一样的。
`, selector)
}

func sendConfirmExportToWA(c tele.Context, sn string, hex string) error {
	selector := &tele.ReplyMarkup{}
	baseUrl, _ := url.JoinPath(msbconf.WebappUrl, "export")
	webAppUrl := fmt.Sprintf("%s?sn=%s&hex=%s", baseUrl, sn, hex)
	log.Debugln("webapp export link is:", webAppUrl)
	webapp := tele.WebApp{URL: webAppUrl}
	btnExport := selector.WebApp("Continue export/继续导出 →", &webapp)
	selector.Inline(selector.Row(btnExport))

	return c.Reply(`
Exporting to WhatsApp requires <a href="https://github.com/star-39/msb_app">Msb App</a> due to their restrictions, then press "Continue export".
导出到 WhatsApp 需要手机上安装<a href="https://github.com/star-39/msb_app">Msb App</a>, 然后按下"继续导出".

Download:下载:
<b>iPhone:</b> AppStore(N/A.暂无), <a href="https://github.com/star-39/msb_app/releases/latest/download/msb_app.ipa">IPA</a>
<b>Android:</b> GooglePlay(N/A.暂无), <a href="https://github.com/star-39/msb_app/releases/latest/download/msb_app.apk">APK</a>
`, tele.ModeHTML, tele.NoPreview, selector)
}

func genSDnMnEInline(canManage bool, isTGS bool, sn string) *tele.ReplyMarkup {
	selector := &tele.ReplyMarkup{}
	btnSingle := selector.Data("Download this sticker/下载这张贴图", CB_DN_SINGLE)
	btnAll := selector.Data("Download sticker set/下载整个贴图包", CB_DN_WHOLE)
	btnMan := selector.Data("Manage sticker set/管理这个贴图包", CB_MANAGE)
	// btnExport := selector.Data("Export to WhatsApp/导出到 WhatsApp", CB_EXPORT_WA)
	if canManage {
		selector.Inline(selector.Row(btnSingle), selector.Row(btnAll), selector.Row(btnMan))
	} else {
		if isTGS {
			//If is TGS, do not support export to WA.
			selector.Inline(selector.Row(btnSingle), selector.Row(btnAll))
		} else {
			selector.Inline(selector.Row(btnSingle), selector.Row(btnAll))
		}
	}
	return selector
}

func sendAskSDownloadChoice(c tele.Context, s *tele.Sticker) error {
	selector := genSDnMnEInline(false, s.Animated, s.SetName)
	return c.Reply(`
You can download this sticker or the whole sticker set, please select below.
你可以下载这个贴图或者其所属的整个贴图包, 请选择:
`, selector)
}

func sendAskSChoice(c tele.Context, sn string) error {
	selector := genSDnMnEInline(true, false, sn)
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
你拥有这个贴图包. 你可以下载或者管理这个贴图包, 请选择:
`, selector)
}

func sendAskTGLinkChoice(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnManu := selector.Data("Download sticker set/下载整个贴图包", CB_DN_WHOLE)
	btnMan := selector.Data("Manage sticker set/管理这个贴图包", CB_MANAGE)
	selector.Inline(selector.Row(btnManu), selector.Row(btnMan))
	return c.Reply(`
You own this sticker set. You can download or manage this sticker set, please select below.
你拥有这个贴图包. 你可以下载或者管理这个贴图包, 请选择:
`, selector)
}

func sendAskWantSDown(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btn1 := selector.Data("Yes", CB_DN_WHOLE)
	btnNo := selector.Data("No", CB_BYE)
	selector.Inline(selector.Row(btn1), selector.Row(btnNo))
	return c.Reply(`
You can download this sticker set. Press Yes to continue.
你可以下载这个贴图包, 按下 Yes 来继续.
`, selector)
}

func sendAskWantImportOrDownload(c tele.Context, avalAsEmoji bool) error {
	msg := ""
	selector := &tele.ReplyMarkup{}
	btnImportSticker := selector.Data("Import as sticker set/作为 Telegram 普通贴图包导入", CB_OK_IMPORT)
	btnImportEmoji := selector.Data("Import as CustomEmoji/作为表情贴导入", CB_OK_IMPORT_EMOJI)
	btnDownload := selector.Data("Download/下载", CB_OK_DN)
	if avalAsEmoji {
		selector.Inline(selector.Row(btnImportSticker), selector.Row(btnImportEmoji), selector.Row(btnDownload))
		msg = `
You can import this sticker set to Telegram or download it.
Import as Custom Emoji is also available, however you will need Telegram Premium to send them.
你可以下载或导入这个贴图包到 Telegram.
也可以作为表情贴导入，但是发送需要 Telegram 会员。`
	} else {
		selector.Inline(selector.Row(btnImportSticker), selector.Row(btnDownload))
		msg = `
You can import this sticker set to Telegram or download it.
你可以下载或导入这个贴图包到 Telegram.`
	}

	return c.Reply(msg, selector)
}

func sendAskWhatToDownload(c tele.Context) error {
	return c.Send("Please send a sticker that you want to download, or its share link(can be either Telegram or LINE ones)\n" +
		"请发送想要下载的贴图, 或者是贴图包的分享链接(可以是Telegram 或 LINE 链接).")
}

func sendAskTitle_Import(c tele.Context) error {
	ld := users.data[c.Sender().ID].lineData
	ld.TitleWg.Wait()
	log.Debug("titles are::")
	log.Debugln(ld.I18nTitles)
	selector := &tele.ReplyMarkup{}

	var titleButtons []tele.Row
	var titleText string
	for i, t := range ld.I18nTitles {
		if t == "" {
			continue
		}
		title := escapeTagMark(t) + " @" + botName
		btn := selector.Data(title, strconv.Itoa(i))
		row := selector.Row(btn)
		titleButtons = append(titleButtons, row)
		titleText = titleText + "\n<code>" + title + "</code>"
	}

	if len(titleButtons) == 0 {
		btnDefault := selector.Data(escapeTagMark(ld.Title)+" @"+botName, CB_DEFAULT_TITLE)
		titleButtons = []tele.Row{selector.Row(btnDefault)}
	}
	selector.Inline(titleButtons...)

	return c.Send("Please send a title for this sticker set. You can also select an original title below:\n"+
		"请发送贴图包的标题.你也可以按下面的按钮自动填上合适的原版标题:\n"+
		titleText, selector, tele.ModeHTML)
}

func sendAskTitle(c tele.Context) error {
	return c.Send("Please send a title for this sticker set.\n" +
		"请发送贴图包的标题.")
}

func sendAskID(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnAuto := selector.Data("Auto Generate/自动生成", "auto")
	selector.Inline(selector.Row(btnAuto))
	return c.Send(`
Please send an ID for sticker set, used in share link.
Can contain alphanum and underscore only.
请设定贴图包的 ID, 用于分享链接。
只可以含有英文、数字、下划线。
For example: 例如:
<code>My_favSticker21</code>

ID is usually not important, you can press Auto Generate.
ID 通常不重要, 你可以按下"自动生成".`, selector, tele.ModeHTML)
}

func sendAskImportLink(c tele.Context) error {
	return c.Send(`
Please send LINE/Kakao store link of the sticker set. You can obtain this link from App by going to sticker store and tapping Share->Copy Link.
请发送贴图包的 LINE/Kakao Store 链接. 你可以在App 里的贴图商店点击右上角的“分享” -> “复制链接”来获取链接.
For example: 例如:
<code>https://store.line.me/stickershop/product/7673/ja</code>
<code>https://e.kakao.com/t/pretty-all-friends</code>
<code>https://emoticon.kakao.com/items/lV6K2fWmU7CpXlHcP9-ysQJx9rg=?referer=share_link</code>
`, tele.ModeHTML)
}

func sendNotifySExist(c tele.Context, lineID string) bool {
	lines := queryLineS(lineID)
	if len(lines) == 0 {
		return false
	}
	message := "This sticker set exists in our database, you can continue import or just use them if you want.\n" +
		"此套贴图包已经存在于数据库中, 你可以继续导入, 或者使用下列现成的贴图包\n\n"

	var entries []string
	for _, l := range lines {
		if l.Ae {
			entries = append(entries, fmt.Sprintf(`<a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title))
		} else {
			// append to top
			entries = append([]string{fmt.Sprintf(`★ <a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title)}, entries...)
		}
	}
	if len(entries) > 5 {
		entries = entries[:5]
	}
	message += strings.Join(entries, "\n")
	c.Send(message, tele.ModeHTML)
	return true
}

func sendSearchResult(entriesWant int, lines []LineStickerQ, c tele.Context) error {
	var entries []string
	message := "Search Results: 搜索结果：\n"

	for _, l := range lines {
		l.Tg_title = strings.TrimSuffix(l.Tg_title, " @"+botName)
		if l.Ae {
			entries = append(entries, fmt.Sprintf(`<a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title))
		} else {
			// append to top
			entries = append([]string{fmt.Sprintf(`★ <a href="%s">%s</a>`, "https://t.me/addstickers/"+l.Tg_id, l.Tg_title)}, entries...)
		}
	}

	if entriesWant == -1 && len(entries) > 120 {
		c.Send("Too many results, please narrow your keyword, truncated to 120 entries.\n" +
			"搜索结果过多，已缩减到 120 个，请使用更准确的搜索关键字。")
		entries = entries[:120]
	}
	if entriesWant != -1 && len(entries) > entriesWant {
		entries = entries[:entriesWant]
	}
	if len(entries) > 30 {
		eChunks := chunkSlice(entries, 30)
		for _, eChunk := range eChunks {
			msgToSend := message + strings.Join(eChunk, "\n")
			c.Send(msgToSend, tele.ModeHTML)
		}
	} else {
		message += strings.Join(entries, "\n")
		c.Send(message, tele.ModeHTML)
	}

	return nil
}

func sendAskStickerFile(c tele.Context) error {
	return c.Send("Please send images/photos/stickers(less than 120 in total),\n" +
		"or send an archive containing image files,\n" +
		"wait until upload complete, then tap 'Done adding'.\n\n" +
		"请发送任意格式的图片、视频或贴图（总数少于 120 张），\n" +
		"或者发送包含贴图文件的压缩包，\n" +
		"等待所有文件上传完成，然后点击「上传完成」。\n")
}

func sendInStateWarning(c tele.Context) error {
	command := users.data[c.Sender().ID].command
	state := users.data[c.Sender().ID].state

	return c.Send(fmt.Sprintf(`
Please send content according to instructions.
请按照 bot 提示发送相应内容.
Current command: %s
Current state: %s
You can also send /quit to terminate session.
你也可以发送 /quit 来中断对话.
`, command, state))
}

func sendNoSessionWarning(c tele.Context) error {
	return c.Send("Please use /start or send LINE/Kakao/Telegram links or stickers.\n请使用 /start 或者发送 LINE/Kakao/Telegram 链接或贴图.")
}

func sendAskSTypeToCreate(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnRegular := selector.Data("Regular sticker set/普通贴图包", CB_REGULAR_STICKER)
	btnCustomEmoji := selector.Data("Custom Emoji/表情贴", CB_CUSTOM_EMOJI)

	selector.Inline(selector.Row(btnRegular), selector.Row(btnCustomEmoji))
	return c.Send("What kind of sticker set you want to create?\nNote that custom emoji can only be sent by Telegram Premium member."+
		"你想要创建哪种类型的贴图包?\n请注意只有 Telegram 会员可以发送表情贴。", selector)
}

func sendAskEmojiAssign(c tele.Context) error {
	ud := udFromCtx(c)
	if ud == nil || ud.stickerData == nil {
		return nil
	}
	sd := ud.stickerData
	if sd.pos < 0 || sd.pos >= len(sd.stickers) {
		log.Errorf("No sticker available for emoji assignment at pos %d, total %d", sd.pos, len(sd.stickers))
		return errNoStickerAvailable
	}
	sf := sd.stickers[sd.pos]
	if sf == nil {
		log.Errorf("Sticker data is nil for emoji assignment at pos %d", sd.pos)
		return errNoStickerAvailable
	}
	total := sd.lAmount
	if total == 0 {
		total = len(sd.stickers)
	}
	waitedForPreview := false
	if sf.fileID == "" && sf.oPath == "" {
		sf.wg.Wait()
		waitedForPreview = true
	}
	if err := sessionContextErr(ud); err != nil {
		return err
	}
	if waitedForPreview && sf.cError != nil {
		return sf.cError
	}
	if sf.fileID == "" && sf.oPath == "" {
		return errors.New("sticker preview file is not ready")
	}
	caption := fmt.Sprintf(`
Send emoji(s) representing this sticker.
请发送代表这个贴图的 emoji（可以多个）.

%d of %d
`, sd.pos+1, total)

	if sf.fileID != "" {
		msg, _ := c.Bot().Send(c.Sender(), &tele.Sticker{
			File: tele.File{FileID: sf.fileID},
		})
		_, err := c.Bot().Reply(msg, caption)
		return err
	}

	err := c.Send(&tele.Video{
		File:    tele.FromDisk(sf.oPath),
		Caption: caption,
	})
	if err != nil {
		err2 := c.Send(&tele.Video{
			File:    tele.FromDisk(sf.oPath),
			Caption: caption,
		})
		if err2 != nil {
			err3 := c.Send(&tele.Document{
				File:     tele.FromDisk(sf.oPath),
				FileName: filepath.Base(sf.oPath),
				Caption:  caption,
			})
			if err3 != nil {
				err4 := c.Send(&tele.Sticker{File: tele.File{FileID: sf.oPath}})
				if err4 != nil {
					return err4
				}
			}
		}
	}
	return nil
}

func sendFatalError(err error, c tele.Context) {
	if c == nil {
		return
	}
	var errMsg string
	if err != nil {
		errMsg = sanitizeErrorText(err)
		errMsg = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(errMsg)
	}

	if isTelegramTemporaryServerError(err) {
		c.Send("<b>Telegram server is temporarily unavailable. Please try again later. /start\n"+
			"Telegram 服务器暂时响应超时或发生错误，请稍后再试。</b>\n\n"+
			"This is usually a temporary Telegram-side issue, not a problem with your sticker link.\n"+
			"这通常是 Telegram 端暂时性问题，不是你的贴图链接错误。\n\n"+
			"<code>"+errMsg+"</code>", tele.ModeHTML, tele.NoPreview)
		return
	}

	c.Send("<b>Fatal error encounterd. Please try again. /start\n"+
		"发生错误，请点击 /start 重试。</b>\n\n"+
		"You can report this error to https://github.com/akira02/chiaki-sticker-bot/issues\n\n"+
		"<code>"+errMsg+"</code>", tele.ModeHTML, tele.NoPreview)
}

func sendExecEmojiAssignFinished(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	msg := fmt.Sprintf(`
LINE Cat: <code>%s</code>
LINE ID: <code>%s</code>
TG ID: <code>%s</code>
TG Title: <a href="%s">%s</a>

Success. 成功完成. /start

If you like this bot, please give us a ⭐️
如果你喜欢这个 Bot，请帮我们点个 ⭐️
https://github.com/akira02/chiaki-sticker-bot
	`, ud.lineData.Category,
		ud.lineData.Id,
		ud.stickerData.id,
		"https://t.me/addstickers/"+ud.stickerData.id,
		escapeTagMark(ud.stickerData.title),
	)
	return c.Send(msg, tele.ModeHTML)
}

// Return:
// string: Text of the message.
// *tele.Message: The pointer of the message.
// error: error
func sendProcessStarted(ud *UserData, c tele.Context, optMsg string) (string, *tele.Message, error) {
	message := fmt.Sprintf(`
Preparing stickers, please wait...
正在准备贴图, 请稍候...

LINE Cat: <code>%s</code>
LINE ID: <code>%s</code>
TG ID: <code>%s</code>
TG TYPE: <code>%s</code>
TG Title: <a href="%s">%s</a>

<b>Progress / 进度</b>
<code>%s</code>
`, ud.lineData.Category,
		ud.lineData.Id,
		ud.stickerData.id,
		ud.stickerData.stickerSetType,
		"https://t.me/addstickers/"+ud.stickerData.id,
		escapeTagMark(ud.stickerData.title),
		optMsg)
	ud.progress = message

	teleMsg, err := c.Bot().Send(c.Recipient(), message, tele.ModeHTML)
	ud.progressMsg = teleMsg
	return message, teleMsg, err
}

// if progressText is empty, a progress bar will be generated based on cur and total.
func editProgressMsg(cur int, total int, progressText string, originalText string, teleMsg *tele.Message, c tele.Context) error {
	defer func() {
		if r := recover(); r != nil {
			log.Errorln("editProgressMsg encountered panic! ignoring...", string(debug.Stack()))
		}
	}()

	header := originalText[:strings.LastIndex(originalText, "<code>")]
	prog := ""

	if progressText != "" {
		prog = progressText
		goto SEND
	}
	cur = cur + 1
	if cur == 1 {
		prog = fmt.Sprintf("<code>[=>                  ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.25)*float64(total)) {
		prog = fmt.Sprintf("<code>[====>               ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.5)*float64(total)) {
		prog = fmt.Sprintf("<code>[=========>          ]\n       %d of %d</code>", cur, total)
	} else if cur == int(float64(0.75)*float64(total)) {
		prog = fmt.Sprintf("<code>[==============>     ]\n       %d of %d</code>", cur, total)
	} else if cur == total {
		prog = fmt.Sprintf("<code>[====================]\n       %d of %d</code>", cur, total)
	} else {
		return nil
	}
SEND:
	messageText := header + prog
	c.Bot().Edit(teleMsg, messageText, tele.ModeHTML)
	return nil
}

func sendAskSToManage(c tele.Context) error {
	if c.Sender().ID == msbconf.AdminUid {
		selector := &tele.ReplyMarkup{}
		btnAdminManage := selector.Data("Admin manage", CB_ADMIN_MANAGE)
		selector.Inline(selector.Row(btnAdminManage))
		return c.Send("Send a sticker from the sticker set that want to edit,\n"+
			"or send its share link.\n\n"+
			"你想要修改哪个贴图包? 请发送那个贴图包内任意一张贴图,\n"+
			"或者是它的分享链接.", selector)
	}
	return c.Send("Send a sticker from the sticker set that want to edit,\n" +
		"or send its share link.\n\n" +
		"你想要修改哪个贴图包? 请发送那个贴图包内任意一张贴图,\n" +
		"或者是它的分享链接.")
}

func sendUserOwnedS(c tele.Context) error {
	return sendStickerSetList(c, c.Sender().ID, "You own following stickers:")
}

func sendAdminManagedS(c tele.Context) error {
	return sendStickerSetList(c, -1, "Admin manageable stickers:")
}

func sendStickerSetList(c tele.Context, uid int64, header string) error {
	usq := queryUserS(uid)
	if usq == nil {
		return errors.New("no sticker owned")
	}

	var entries []string

	for _, us := range usq {
		date := time.Unix(us.timestamp, 0).Format("2006-01-02 15:04")
		title := strings.TrimSuffix(us.tg_title, " @"+botName)
		//workaround for empty title.
		if title == "" || title == " " {
			title = "_"
		}
		entry := fmt.Sprintf(`<a href="https://t.me/addstickers/%s">%s</a>`, us.tg_id, title)
		entry += " | " + date
		entries = append(entries, entry)
	}

	if len(entries) > 30 {
		eChunks := chunkSlice(entries, 30)
		for _, eChunk := range eChunks {
			message := header + "\n"
			message += strings.Join(eChunk, "\n")
			c.Send(message, tele.ModeHTML)
		}
	} else {
		message := header + "\n"
		message += strings.Join(entries, "\n")
		c.Send(message, tele.ModeHTML)
	}
	return nil
}

func sendAskEditChoice(c tele.Context) error {
	ud := users.data[c.Sender().ID]
	selector := &tele.ReplyMarkup{}
	btnAdd := selector.Data("Add sticker/添加贴图", CB_ADD_STICKER)
	btnDel := selector.Data("Delete sticker/删除贴图", CB_DELETE_STICKER)
	btnDelset := selector.Data("Delete sticker set/删除贴图包", CB_DELETE_STICKER_SET)
	btnChangeTitle := selector.Data("Change title/修改标题", CB_CHANGE_TITLE)
	btnExit := selector.Data("Exit/退出", "bye")

	if msbconf.WebappUrl != "" {
		baseUrl, _ := url.JoinPath(msbconf.WebappUrl, "edit")
		url := fmt.Sprintf("%s?ss=%s&dt=%d",
			baseUrl,
			ud.stickerData.id,
			time.Now().Unix())
		log.Debugln("WebApp URL is : ", url)
		webApp := &tele.WebApp{
			URL: url,
		}
		btnEdit := selector.WebApp("Change order or emoji/修改顺序或 Emoji", webApp)
		selector.Inline(
			selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnEdit), selector.Row(btnChangeTitle), selector.Row(btnExit))
	} else {
		selector.Inline(
			selector.Row(btnAdd), selector.Row(btnDel), selector.Row(btnDelset), selector.Row(btnChangeTitle), selector.Row(btnExit))
	}

	return c.Send(fmt.Sprintf(`
ID: <code>%s</code>
Title: <a href="https://t.me/addstickers/%s">%s</a>

What do you want to edit? Please select below:
你想要修改贴图包的什么内容? 请选择:`,
		users.data[c.Sender().ID].stickerData.id,
		ud.stickerData.id,
		ud.stickerData.title),
		selector, tele.ModeHTML)
}

func sendAskSDel(c tele.Context) error {
	return c.Send("Which sticker do you want to delete? Please send it.\n" +
		"你想要删除哪一个贴图? 请发送那个贴图")
}

func sendConfirmDelset(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnYes := selector.Data("Yes", CB_YES)
	btnNo := selector.Data("No", CB_NO)
	selector.Inline(selector.Row(btnYes), selector.Row(btnNo))

	return c.Send("You are attempting to delete the whole sticker set, please confirm.\n"+
		"你将要删除整个贴图包, 请确认.", selector)
}

func sendSFromSS(c tele.Context, ssid string, reply *tele.Message) error {
	ss, _ := c.Bot().StickerSet(ssid)
	if reply != nil {
		c.Bot().Reply(reply, &ss.Stickers[0])
	} else {
		c.Send(&ss.Stickers[0])
	}
	return nil
}

func sendFLWarning(c tele.Context) error {
	return c.Send(`
It might take longer to process this sticker set (2-8 minutes)... 
This warning indicates that you might triggered Telegram's flood limit, and bot is trying to re-submit.
Due to this mechanism, resulted sticker set might contains duplicate or missing sticker, please check manually after done.

此贴图包可能需要更长时间处理(2-8 分钟)...
看到这一条警告表示Telegram 可能限制了你创建贴图包的频率, 且 bot 正在自动尝试重新制作, 因此得出的贴图包可能会重复或缺失贴图, 请在完成制作后再检查一下.
`)
}

func sendTooManyFloodLimits(c tele.Context) error {
	return c.Send("Sorry, it seems that you have triggered Telegram's flood limit for too many times, it's recommended try again after a while.\n" +
		"抱歉, 你似乎触发了Telegram 的贴图制作次数限制, 建议你过一段时间后再试一次.")
}

func sendNoCbWarn(c tele.Context) error {
	return c.Send("Please press a button! /quit\n请选择按钮!")
}

func sendBadIDWarn(c tele.Context) error {
	return c.Send(`
Bad ID. try again or press Auto Generate. /quit
Can contain alphanum and underscore only, must begin with alphabet, must not contain consecutive underscores.
只可以含有英文、数字、下划线, 必须以英文字母开头，不可以有连续下划线.
ID 错误, 请再试一次或按下'自动生成'按钮. /quit`)
}

func sendIDOccupiedWarn(c tele.Context) error {
	return c.Send("ID already occupied! try another one. ID 已经被占用, 请试试另一个.")
}

func sendBadImportLinkWarn(c tele.Context) error {
	return c.Send("Invalid import link, make sure its a LINE Store link or Kakao store link. Try again or /quit\n"+
		"无效的链接, 请查看是否为LINE 贴图商店的链接, 或是Kakao emoticon的链接\n\n"+
		"For example: 例如:\n"+
		"<code>https://store.line.me/stickershop/product/7673/ja</code>\n"+
		"<code>https://e.kakao.com/t/pretty-all-friends</code>", tele.ModeHTML)
}

func sendNoSToManage(c tele.Context) error {
	return c.Send("Sorry, you have not created any sticker set yet. You can use /import or /create .\n" +
		"抱歉, 你还未创建过贴图包, 你可以使用 /create 或 /import 来创建贴图包")
}

func sendPromptStopAdding(c tele.Context) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/上传完成", CB_DONE_ADDING)
	selector.Inline(selector.Row(btnDone))
	return c.Send("Continue sending files or press button below to stop adding.\n"+
		"请继续发送文件，或者点击下方按钮完成上传。", selector)
}

func replySFileOK(c tele.Context, count int) error {
	selector := &tele.ReplyMarkup{}
	btnDone := selector.Data("Done adding/上传完成", CB_DONE_ADDING)
	selector.Inline(selector.Row(btnDone))
	return c.Reply(
		fmt.Sprintf("File OK. Got %d stickers. Continue sending files or press button below to stop adding.\n"+
			"文件正常，已收到 %d 张贴图。请继续发送文件，或者点击下方按钮完成上传。", count, count), selector)
}

func sendSEditOK(c tele.Context) error {
	return c.Send(
		"Successfully edited sticker set. /start\n" +
			"成功修改贴图包. /start")
}

func sendStickerSetFullWarning(c tele.Context) error {
	return c.Send(
		"Warning: Your sticker set is already full. You cannot add new sticker.\n" +
			"提示：当前贴图包已满，无法添加新的贴图。")
}

// func sendEditingEmoji(c tele.Context) error {
// 	return c.Send("Commiting changes...\n正在套用变更，请稍候...")
// }

func sendAskSearchKeyword(c tele.Context) error {
	return c.Send("Please send a word that you want to search\n请发送想要搜索的内容")
}

func sendSearchNoResult(c tele.Context) error {
	message := "Sorry, no result.\n没有结果"
	if c.Chat().Type == tele.ChatPrivate {
		message += "\nTry again or /quit\n请试试别的关键字或 /quit"
	}
	return c.Send(message)
}

func sendNotifyNoSessionSearch(c tele.Context) error {
	return c.Send("Here are some search results, use /search to dig deeper or /start to see available commands.\n" +
		"这些是贴图包搜索结果，使用 /search 详细搜索或 /start 来看看可用的指令。")
}

func sendUnsupportedCommandForGroup(c tele.Context) error {
	return c.Send("This command is not supported in group chat, please chat with bot directly.\n" +
		"此指令无法在群组内使用，请与 bot 直接私聊.")
}

func sendBadSearchKeyword(c tele.Context) error {
	return c.Send(fmt.Sprintf(`
Please specify keyword
请指定搜索关键字.

Example: 例如:
/search@%s keyword1 keyword2 ...
/search@%s nekomimi mia
`, botName, botName))
}

func sendPreferKakaoShareLinkWarning(c tele.Context) error {
	msg := `
The link you sent is a Kakao store link.
Use a share link for improved image quality and animated sticker support,
you can obtain it from KakaoTalk app by tapping share->copy link in sticker store.

你发送的是 Kakao 商店的链接.
使用分享链接才能支持动态贴图, 静态贴图的画质也更高。
你可以在KakaoTalk App 内的贴图商店点击“分享” -> “复制链接”来获取分享链接。

eg:例如: <code>https://emoticon.kakao.com/items/lV6K2fWmU7CpXlHcP9-ysQJx9rg=?referer=share_link</code>
`
	err := c.Reply(&tele.Photo{
		File:    tele.File{FileID: FID_KAKAO_SHARE_LINK},
		Caption: msg,
	}, tele.ModeHTML)
	if err != nil {
		c.Reply(msg, tele.ModeHTML)
	}
	return nil
}

func sendUseCommandToImport(c tele.Context) error {
	return c.Send("Please use /create to create sticker set using your own photos and videos. /start\n" +
		"请使用 /create 指令，通过自己的图片和视频创建贴图包。/start")
}

func sendOneStickerFailedToAdd(c tele.Context, pos int, err error) error {
	if errors.Is(err, msbimport.ErrStickerTooLarge) {
		return c.Reply(stickerCompressionFailedMessage(err), tele.NoPreview)
	}
	return c.Reply(fmt.Sprintf(`
Failed to add one sticker.
一张贴图添加失败
Index: %d
Error: %s
`, pos, sanitizeErrorText(err)))
}

func sendStickerCompressionFailed(c tele.Context, err error) error {
	if c == nil {
		return nil
	}
	return c.Send(stickerCompressionFailedMessage(err), tele.NoPreview)
}

func stickerCompressionFailedMessage(err error) string {
	reason := ""
	if err != nil {
		reason = sanitizeErrorText(err)
	}
	return "This sticker is too large or too complex to fit Telegram's video sticker limit, even after reducing bitrate and shortening the animation.\n" +
		"这张贴图太大或动画太复杂；即使降低 bitrate 并缩短秒数后，仍无法压到 Telegram 视频贴图限制内。\n\n" +
		"Reason: " + reason + "\n" +
		"原因：" + reason + "\n\n" +
		"If this sticker should work, please report it here:\n" +
		"如果这张贴图理论上可以导入，请到这里反馈：\n" +
		"https://github.com/akira02/chiaki-sticker-bot/issues"
}

func sendBadSNWarn(c tele.Context) error {
	return c.Reply("Wrong sticker or link!\n贴图或链接错误!")
}

func sendSSTitleChanged(c tele.Context) error {
	msg := `
Successfully changed title.
新标题设置完成`
	return c.Reply(msg, tele.ModeHTML)
}
func sendSSTitleFailedToChanged(c tele.Context) error {
	msg := `
Failed to change title, please try again.
新标题设置失败，请再试一次。`
	return c.Reply(msg, tele.ModeHTML)
}

// func sendInvalidEmojiWarn(c tele.Context) error {
// 	return c.Reply(`
// Sorry, this emoji is invalid, it has been defaulted to ⭐️, you can edit it after done by using /manage command.
// 抱歉，这个 emoji 无效，并且已默认设定为⭐️，你可以在完成制作后使用 /manage 来修改。
// 	`)
// }

func sendProcessingStickers(c tele.Context) error {
	return c.Send(`
Processing stickers, please wait a while...
正在制作贴图，请稍等...
`)
}
