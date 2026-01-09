args = cli.Json(
    "kv", 
    cli.setJsonSchema(
        <<<JSON
{
  "type": "object",
  "properties": {
    "kvs": {
      "type": "array",
      "title": "配置认证信息",
      "minItems": 1,
      "items": {
        "properties": {
          "kind": {
            "title": "认证类型",
            "type": "string",
            "enum": [
                "Header",
                "Cookie"
            ],
            "default": "Header",
            "description": "设置认证信息，如果需要对整体 Cookie 进行测试，请使用 Header 模式，Key 填写 Cookie 即可"
            },
          "force": {
            "type": "boolean",
            "default": true,
            "title": "强制添加",
            "description":"默认如果源数据包不存在这个认证信息，则不测试，勾选本选项后，即使源数据包不存在这个认证信息，也会测试"
          },
          "key": {
            "type": "string",
            "title": "Key",
            "description": "认证字段名，如果需要测试 Cookie，仅填写 Cookie Key 即可，例如 PHPSESSION 或 JSESSION 等"
          },
          "value": {
            "type": "string",
            "title": "Value",
            "description": "认证字段值（单行），如果需要多个认证，请添加新的认证信息，如果有多个关联的认证信息，请设置分组"
          },
          "group": {
            "type": "string",
            "title": "关联分组",
            "description": "关联分组，用于关联多个认证信息，同一个分组ID的认证信息会同时被添加测试，不同分组ID的认证信息会分别测试",
            "enum": [
              "",
              "1",
              "2",
              "3",
              "4",
              "5",
              "6",
              "7",
              "8",
              "9",
              "10"
            ]
          }
        },
        "required": [
          "key",
          "value",
          "kind",
          "force"
        ]
      }
    }
  }
}
JSON, 
        cli.setUISchema(
            cli.uiGroups(cli.uiGroup(cli.uiField(
                "kvs", 
                1, 
                cli.uiFieldWidget(cli.uiWidgetTable), 
                cli.uiFieldGroups(cli.uiGroup(cli.uiField(
                    "items", 
                    1, 
                    cli.uiFieldGroups(cli.uiGroup(
                        cli.uiTableField("kind", 100),
                        cli.uiTableField("key", 100),
                        cli.uiTableField("value", 260, cli.uiFieldWidget(cli.uiWidgetTextarea)),
                        cli.uiTableField("force", 86),
                        cli.uiTableField("group", 86)
                    )), 
                ))), 
            ))),
            cli.uiGlobalFieldPosition(cli.uiPosHorizontal),
        ), 
    ), 
    cli.setRequired(true), 
)
enableUnauth = cli.Bool("enable-unauth", cli.setRequired(true), cli.setVerboseName("未授权检测"),cli.setHelp("当用户输入认证Value不为空时勾选此开关,开启未授权检测"), cli.setDefault(true))
disableDomain = cli.String("disable-domain", cli.setRequired(false), cli.setVerboseName("不检测的域名"), cli.setCliGroup("高级（可选参数）"))
disablePath = cli.String("disable-path", cli.setRequired(false), cli.setVerboseName("不检测的路径"), cli.setCliGroup("高级（可选参数）"))
disableTypes := cli.StringSlice("disable-types", cli.setMultipleSelect(true), cli.setSelectOption("js", "js"),cli.setSelectOption("css", "css"),cli.setSelectOption("img", "img"),cli.setSelectOption("font", "font"),cli.setSelectOption("static", "static"),cli.setDefault("js,css,img,font"), cli.setRequired(false), cli.setVerboseName("忽略静态资源类型"), cli.setCliGroup("高级（可选参数）"))

enableDomain = cli.String("enable-domain", cli.setRequired(false), cli.setVerboseName("需检测域名"), cli.setCliGroup("高级（可选参数）"))
enablePath = cli.String("enable-path", cli.setRequired(false), cli.setVerboseName("需检测路径"), cli.setCliGroup("高级（可选参数）"))
enableResponseKeyword = cli.Text("enable-response-content", cli.setRequired(false), cli.setVerboseName("需检测响应内容标志值"), cli.setCliGroup("高级（可选参数）"))
enableResponseKeywordRegexp = cli.Bool(
    "enable-response-content-regexp", 
    cli.setRequired(false), 
    cli.setVerboseName("响应内容标志值开启正则"), 
    cli.setCliGroup("高级（可选参数）"), 
)
cli.check()

yakit.AutoInitYakit()

ruleMap := {
    "js": [
        "*.js","*.mjs","*.jsx","*.ts","*.tsx","*.map",
    ],
    "css": [
        "*.css","*.scss","*.sass","*.less",
    ],
    "img": [
        "*.png","*.jpg","*.jpeg","*.gif","*.svg","*.webp","*.ico","*.bmp",
    ],
    "font": [
        "*.woff","*.woff2","*.ttf","*.otf","*.eot",
    ]
}

once = sync.NewOnce()
groupMap = make(map[string][]any)
disableRules := []string{}

for _, item = range args["kvs"] {
    group = item["group"]
    if group == "" {
        group = uuid()
    }
    if group not in groupMap {
        groupMap[group] = make([]any, 0, 1)
    }
    groupMap[group] = append(groupMap[group], item)
}

for _, types := range disableTypes{
    disableRules = append(disableRules, ruleMap[types]...)
}


func product(sli, curr) {
    ret = make([][]any, 0, len(sli)*len(curr))
    for i in len(sli) {
        for j in len(curr) {
            tmp := make([]any, len(sli[i]))
            for n in len(sli[i]) {
                tmp[n] = sli[i][n]
            }
            ret.Append(append(tmp, curr[j]))
        }
    }
    return ret
}

func Product(sli) {
    if len(sli) == 0 {
        return sli
    }
    ret = make([][]any, 0, len(sli[0]))
    for i in len(sli[0]) {
        ret = append(ret, [sli[0][i]])
    }
    for i =1;i<len(sli);i++ {
        ret = product(ret, sli[i])
    }
    return ret
}

handleReq = (https, reqBytes, baseResponse, handleTag) => {
    // yakit.Code(reqBytes)
    poc.HTTP(
        reqBytes, 
        poc.https(https), 
        poc.saveHandler(response => {
            tag = ""

            // ----------- 是否未授权检测 -----------
            isUnauth := str.Contains(handleTag, "未授权检测")

            if len(enableResponseKeywordList) > 0 {
                if respMatch(response.RawPacket, enableResponseKeywordList...) {
                    tag = "响应内容标志值匹配"
                    response.Red()
                } else {
                    tag = "响应内容标志值消失"
                    response.Green()
                }
            } else {

                // 只比较 Response Body，避免因为响应body过小，而导致重复度极高的情况
                _, baseBody := str.SplitHTTPHeadersAndBodyFromPacket(baseResponse)
                _, rspBody := str.SplitHTTPHeadersAndBodyFromPacket(response.RawPacket)

                sim := str.CalcSimilarity(baseBody, rspBody)

                // 对于未授权检测，如果对比重复度小于 0.8 不予展示
                if isUnauth && sim < 0.8 {
                    return    
                }

                if sim > 0.9 {
                    response.Red()
                } elif sim <= 0.4 {
                    response.Green()
                } else {
                    response.Grey()
                }

                showSim = "%.2f" % (sim * 100.0)
                tag = f"相似：${showSim}% "
            }
            response.AddTag(tag)
            response.AddTag(handleTag)
        }), 
    )
}

mirrorFilteredHTTPFlow = (https, url, req, rsp, body) => {
    replaceResults = []
    emptyResults = []
    originReq = req

    // ------------------------------
    // 判断当前是否存在认证信息（决定是否启用 未授权测试）
    // ------------------------------
    hasAuth := false
    for _, items := range groupMap {
        for _, item := range items {
            if str.TrimSpace(item.value) != "" {
                hasAuth = true
                break
            }
        }
    }

    for _, items = range groupMap {
        replacedTmp = []
        emptyTmp  = []
        for _, item = range items {
            key = item.key
            values = str.ParseStringToLines(item.value)
            kind = item.kind
            force = item.force

            skipped = false
            host, _, _ = str.ParseStringToHostPort(url)
            path = poc.GetHTTPRequestPath(req)

            if disableDomain != "" {
                rules = disableDomain.Split(",").Map(i => i.Trim())
                skipped = str.MatchAnyOfGlob(host, rules...)
            }
            

            if !skipped && disablePath != "" {
                skipped = str.MatchAnyOfGlob(path, disablePath.Split(",").Map(i => i.Trim())...)
            }

            if !skipped && disableTypes != nil {
                skipped = str.MatchAnyOfGlob(str.Split(path, "?")[0], disableRules...)
            }


            if !skipped && enableDomain {
                skipped = !str.MatchAnyOfGlob(host, enableDomain.Split(",").Map(i => i.Trim())...)
            }
            

            if !skipped && enablePath {
                skipped = !str.MatchAnyOfGlob(path, enablePath.Split(",").Map(i => i.Trim())...)
            }


            respMatch = str.MatchAnyOfRegexp
            if !enableResponseKeywordRegexp {
                respMatch = str.MatchAnyOfSubString
            }


            enableResponseKeywordList = []
            if !skipped && enableResponseKeyword {
                enableResponseKeywordList = enableResponseKeyword.Split("\n")
                skipped = !respMatch(rsp, enableResponseKeywordList...)
            }
            

            if skipped {
                return
            }

            
            if !force {
                if kind == "Cookie" {
                    if poc.GetHTTPPacketCookie(req, key) == "" {
                        continue
                    }
                } elif kind == "Header" {
                    if poc.GetHTTPPacketHeader(req, key) == "" {
                        continue
                    }
                }
            }

            if kind == "Header" || kind == "Cookie" {
                for v in values {
                    replacedTmp.Append({"key": key, "kind": kind, "value": v})
                }
                emptyTmp.Append({"key": key, "kind": kind, "value": ""})
            }
        }
        if len(replacedTmp) == 0 {
            replacedTmp.Append({"key": "", "kind": "", "value": ""})
            emptyResults.Append({"key": "", "kind": "", "value": ""})
        } else {
            replaceResults.Append(replacedTmp)
            emptyResults.Append(emptyTmp)
        }
    }
    dump(replaceResults)

    for seq in Product(replaceResults) {
        replaceReq = originReq
        tags = []
        for m in seq {
            kind, key, value = m.kind, m.key, m.value
            if kind == "Header" {
                replaceReq = poc.ReplaceHTTPPacketHeader(replaceReq, key, value)
            } else if kind == "Cookie" {
                replaceReq = poc.ReplaceHTTPPacketCookie(replaceReq, key, value)
            }
            tags = append(tags, f"${kind}[${key}: ${value}]")
        }
        tag = f"值 ${str.Join(tags, ", ")}"
        handleReq(https, replaceReq, rsp, tag)
    }
    
    // 未授权测试：仅满足条件时才执行触发条件：
    // 1️. 用户填写了认证
    // 2️. 勾选未授权检测
    if hasAuth && enableUnauth {
        for seq in Product(emptyResults) {
            emptyReq = originReq
            tags = []
            for m in seq {
                kind, key, value = m.kind, m.key, m.value
                if kind == "Header" {
                    if poc.GetHTTPPacketHeader(emptyReq, key) {
                        emptyReq = poc.ReplaceHTTPPacketHeader(emptyReq, key, value)
                    }
                } else if kind == "Cookie" {
                    if poc.GetHTTPPacketCookie(emptyReq, key) {
                        emptyReq = poc.ReplaceHTTPPacketCookie(emptyReq, key, value)
                    }
                }
                tags = append(tags, f"${kind}[${key}]")
            }
            tag = f"未授权检测 | 移除 ${str.Join(tags, ", ")}"
            handleReq(https, emptyReq, rsp, tag)
        }
    }
}