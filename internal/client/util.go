//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

package client

import (
	"net/url"
	"strings"
)

func truncateToken(token string) string {
	if len(token) <= 4 {
		return token
	}
	return "..." + token[len(token)-4:]
}

func sanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	modified := false
	for k := range q {
		lowerK := strings.ToLower(k)
		if strings.Contains(lowerK, "token") || lowerK == "id" || lowerK == "sessionid" {
			val := q.Get(k)
			if len(val) > 4 {
				q.Set(k, "..."+val[len(val)-4:])
				modified = true
			}
		}
	}
	if modified {
		u.RawQuery = q.Encode()
		return u.String()
	}
	return rawURL
}
