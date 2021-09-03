/*-
 * Copyright (c) 2021 Ryo ONODERA <ryo@tetera.org>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY RYO ONODERA AND CONTRIBUTORS
 * ``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
 * TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED.  IN NO EVENT SHALL THE FOUNDATION OR CONTRIBUTORS
 * BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/eduncan911/podcast"
)

type String string
type Strings []String

func cvtDuration(sClkDur string) int64 {
	clkLayout := "15:04:05"

	t, err := time.Parse(clkLayout, sClkDur)
	if err != nil {
		return int64(0)
	}

	h, m, s := t.Clock()
	d := (time.Duration(h) * time.Hour +
		time.Duration(m) * time.Minute +
		time.Duration(s) * time.Second) / time.Microsecond / time.Millisecond

	return int64(d)
}

func createCombinedRSS(sText string, sMp3 string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	timeFmt := "Mon, 02 Jan 2006 15:04:05 +0000"

	defer cancel()

	fpText := gofeed.NewParser()
	feedText, _ := fpText.ParseURLWithContext(sText, ctx)
	fpMp3 := gofeed.NewParser()
	feedMp3, _ := fpMp3.ParseURLWithContext(sMp3, ctx)

	createdDateTime, err := time.Parse(timeFmt, feedText.Updated)
	if err != nil {
		return "Error to parse updated time"
	}

	p := podcast.New(feedText.Title,
			 feedText.Link,
			 feedText.Description,
			 &createdDateTime,
			 &createdDateTime)
	p.Language = "en"
	p.AddAtomLink(feedText.Link)
	p.AddImage(feedMp3.Image.URL)

	for _, itemText := range feedText.Items {
		if len(itemText.Description) == 0 {
			// MP3 case
			continue
		}
		for _, itemMp3 := range feedMp3.Items {
			if (
			   strings.HasPrefix(strings.Replace(itemMp3.Title, "’", "'", -1), strings.Replace(itemText.Title, "’", "'", -1))) {
				item := podcast.Item {
					Title: 		itemText.Title,
					Description:	itemText.Description,
					Link:		itemMp3.Link,
				}

				item.AddEnclosure(
					itemMp3.Enclosures[0].URL,
					podcast.MP3,
					0)

				dur := cvtDuration(itemMp3.ITunesExt.Duration)
				item.AddDuration(dur)

				pubDate := itemMp3.Published
				tPubDate, _ := time.Parse(time.RFC1123Z, pubDate)
				item.AddPubDate(&tPubDate)

				_, err = p.AddItem(item)
				if err != nil {
					//return "ERROR to compose RSS: " + err.Error()
				}
			}
		}
	}

	rss := p.String()
	return rss
}

func (s Strings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, createCombinedRSS(string(s[0]), string(s[1])))
}

func main() {
	urls := Strings{"https://learningenglish.voanews.com/api/zmg_pe$myp", "https://learningenglish.voanews.com/podcast/?zoneId=1579"}
	http.Handle("/rss", urls)
	http.ListenAndServe(":8080", nil)
}
