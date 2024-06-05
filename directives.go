package zui

import (
	"errors"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

func (me *zui2js) doDirectiveAttr(attr *html.Attribute, jsVarNameCurNode string, jsAttrValFnName string, jsPart *htmlTextAndExprsSplitItem) (addAttrs []html.Attribute, err error) {
	const pref = "\n    "
	attr_name := strTrim(attr.Key)
	assert(attr_name != "")
	switch {

	case strings.HasPrefix(attr_name, "on:"):
		parts := strings.Split(strTrim(attr_name[len("on:"):]), "|")
		evt_name := strTrim(parts[0])
		if evt_name == "" {
			return nil, errors.New(me.zuiFilePath + ": event name missing after `on:`")
		}
		evt_fwd, evt_mods := (jsAttrValFnName == ""), parts[1:]
		if evt_fwd {
			jsAttrValFnName = me.nextFnName()
			me.WriteString(pref + "const " + jsAttrValFnName + " = (() => ((evt) => {")
			me.WriteString(pref + "  this.dispatch('" + evt_name + "', evt.detail);")
			me.WriteString(pref + "}));")
		}
		if len(evt_mods) > 0 {
			name_fn := me.nextFnName()
			me.WriteString(pref + "const " + name_fn + " = (() => ((evt) => {")
			if slices.Contains(evt_mods, "trusted") {
				me.WriteString(pref + "  if (!evt.isTrusted) { return; }")
			}
			if slices.Contains(evt_mods, "self") {
				me.WriteString(pref + "  if (!evt.target !== this) { return; }")
			}
			if slices.Contains(evt_mods, "preventDefault") {
				me.WriteString(pref + "  evt.preventDefault();")
			}
			if slices.Contains(evt_mods, "stopPropagation") {
				me.WriteString(pref + "  evt.stopPropagation();")
			}
			me.WriteString(pref + "  " + jsAttrValFnName + "().bind(this)(evt);")
			me.WriteString(pref + "})).bind(this);")
			jsAttrValFnName = name_fn
		}
		me.WriteString(pref + jsVarNameCurNode + ".addEventListener('" + evt_name +
			"', ((evt) => (" + jsAttrValFnName + ")().bind(this)(evt)).bind(this), {" +
			strings.TrimSuffix(
				ıf(!slices.Contains(evt_mods, "once"), "", "once:true,")+
					ıf(!slices.Contains(evt_mods, "passive"), "", "passive:true,")+
					ıf(!slices.Contains(evt_mods, "nonpassive"), "", "passive:false,")+
					ıf(!slices.Contains(evt_mods, "capture"), "", "capture:true,"),
				",") + "});")

	case strings.HasPrefix(attr_name, "bind:"):
		if jsPart == nil || jsPart.jsExpr == nil {
			return nil, errors.New(me.zuiFilePath + ": invalid 'bind' argument `" + attr.Val + "`")
		}
		js_expr_frag := strings.TrimSuffix(jsString(jsPart.jsExpr), ";")
		prop_name := strTrim(attr_name[len("bind:"):])
		var evt_name string
		switch prop_name {
		case "value":
			evt_name = "input"
			me.WriteString(pref + jsVarNameCurNode + ".addEventListener('" + evt_name + "', ((evt) => { " + js_expr_frag + " = " + jsVarNameCurNode + "." + prop_name + "; }).bind(this));")
			addAttrs = append(addAttrs, html.Attribute{Key: prop_name, Val: attr.Val})
		default:
			panic("TODO: implement event-handling for capturing '" + prop_name + "' changes")
		}

	default:
		return nil, errors.New(me.zuiFilePath + ": unknown directive '" + attr_name + "'")
	}
	return
}
