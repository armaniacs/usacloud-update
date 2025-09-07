package transform

import "strings"

func GeneratedHeader() string {
	return "# Updated for usacloud v1.1 by usacloud-update — DO NOT EDIT ABOVE THIS LINE"
}

func DefaultRules() []Rule {
	var rules []Rule

	// 1) 出力タイプcsv/tsvの廃止 -> jsonへ (usacloud文脈に限定)
	rules = append(rules, mk(
		"output-type-csv-tsv",
		`(?i)\busacloud\s+[^\s]*\s+.*?(--output-type|\s-o)\s*=?\s*(csv|tsv)`,
		func(m []string) string { return strings.Replace(m[0], m[2], "json", 1) },
		"v1.0でcsv/tsvは廃止。jsonに置換し、必要なら --query/jq を利用してください",
		"https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	))

	// 2) --selector の廃止 -> 引数へ
	rules = append(rules, mk(
		"selector-to-arg",
		`--selector\s+([^\\s]+)`,
		func(m []string) string {
			// サブケース: name=xxx / id=xxx / tag=xxx などの最右辺を引数へ移行
			kv := m[1]
			val := kv
			if i := strings.Index(kv, "="); i > 0 {
				val = kv[i+1:]
			}
			return strings.Replace(m[0], m[0], val, 1)
		},
		"--selectorはv1で廃止。ID/名称/タグをコマンド引数に指定する仕様へ移行",
		"https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	))

	// 3) リソース名の変更: iso-image -> cdrom
	rules = append(rules, mk(
		"iso-image-to-cdrom",
		`\busacloud\s+iso-image\b`,
		func(m []string) string { return strings.Replace(m[0], "iso-image", "cdrom", 1) },
		"v1ではリソース名がcdromに統一",
		"https://manual.sakura.ad.jp/cloud-api/1.1/cdrom/index.html",
	))

	// 4) リソース名の変更: startup-script -> note
	rules = append(rules, mk(
		"startup-script-to-note",
		`\busacloud\s+startup-script\b`,
		func(m []string) string { return strings.Replace(m[0], "startup-script", "note", 1) },
		"v1ではstartup-scriptはnoteに統一",
		"https://docs.usacloud.jp/usacloud/",
	))

	// 5) リソース名の変更: ipv4 -> ipaddress
	rules = append(rules, mk(
		"ipv4-to-ipaddress",
		`\busacloud\s+ipv4\b`,
		func(m []string) string { return strings.Replace(m[0], "ipv4", "ipaddress", 1) },
		"v1ではIPv4関連はipaddressに整理",
		"https://docs.usacloud.jp/usacloud/references/ipaddress/",
	))

	// 6) product-* -> *-plan (v0系の別名整理)
	for _, pair := range [][2]string{{"product-disk", "disk-plan"}, {"product-internet", "internet-plan"}, {"product-server", "server-plan"}} {
		old, new := pair[0], pair[1]
		rules = append(rules, mk(
			"product-alias-"+old,
			`\busacloud\s+`+old+`\b`,
			func(m []string) string { return strings.Replace(m[0], old, new, 1) },
			"v1系では *-plan へ名称統一",
			"https://docs.usacloud.jp/usacloud/",
		))
	}

	// 7) summary の廃止 -> コメントアウト(手動対応)
	rules = append(rules, mk(
		"summary-removed",
		`^\s*usacloud\s+summary\b.*$`,
		func(m []string) string { return "# " + m[0] },
		"summaryコマンドはv1で廃止。要件に応じて bill/self/各list か rest を利用してください",
		"https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	))

	// 8) object-storageサブコマンドの非サポート(v1方針)
	for _, alias := range []string{"object-storage", "ojs"} {
		rules = append(rules, mk(
			"object-storage-removed-"+alias,
			`^\s*usacloud\s+`+alias+`\b.*$`,
			func(m []string) string { return "# " + m[0] },
			"v1ではオブジェクトストレージ操作は非対応方針。S3互換ツール/他プロバイダやTerraformを検討",
			"https://github.com/sacloud/usacloud/issues/585",
		))
	}

	// 9) --zone all の有効化: 変換は不要だが誤記修正(=の周辺空白) (usacloud文脈に限定)
	rules = append(rules, mk(
		"zone-all-normalize",
		`(\busacloud\s+[^\s]*\s+.*?)--zone\s*=\s*all`,
		func(m []string) string {
			return m[1] + "--zone=all"
		},
		"全ゾーン一括操作は --zone=all を推奨",
		"https://docs.usacloud.jp/usacloud/upgrade/v1_0_0/",
	))

	return rules
}
