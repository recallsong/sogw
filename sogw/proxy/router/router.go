/*
	改写自 https://github.com/labstack/echo 的路由算法和测试代码
	1、将路由的结果从HandlerFunc修改为Destination类型，这样更通用
	2、将Find结果修改Result类型
	3、在Router上定义maxParamNum 和 GetMaxParamNum 和 NewResult
*/
package router

import "net/http"

var Methods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodOptions,
	http.MethodHead,
	http.MethodConnect,
	http.MethodTrace,
	MethodPropfind,
}

type (
	Router struct {
		tree        *node
		maxParamNum int
	}
	node struct {
		kind        kind
		label       byte
		prefix      string
		parent      *node
		children    children
		pnames      []string
		ppath       string
		methodDests *methodDests
	}
	kind        uint8
	children    []*node
	Destination interface{}
	methodDests struct {
		connect  Destination
		delete   Destination
		get      Destination
		head     Destination
		options  Destination
		patch    Destination
		post     Destination
		propfind Destination
		put      Destination
		trace    Destination
	}
	Result struct {
		Pattern        string
		PathParams     []string
		PathValues     []string
		MethodNotAllow bool
		Dest           Destination
	}
)

const (
	skind kind = iota
	pkind
	akind
)

// HTTP methods that not include http package
const (
	MethodPropfind = "PROPFIND"
)

// New returns a new Router instance.
func New() *Router {
	return &Router{
		tree: &node{methodDests: new(methodDests)},
	}
}

func (r *Router) GetMaxParamNum() int {
	return r.maxParamNum
}

func (r *Router) Add(method, path string, dest Destination) {
	// Validate path
	if path == "" {
		panic("router: path cannot be empty")
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/' && path[i] != '.'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], dest, pkind, ppath, pnames)
				return
			}
			r.insert(method, path[:i], nil, pkind, ppath, pnames)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], dest, akind, ppath, pnames)
			return
		}
	}
	r.insert(method, path, dest, skind, ppath, pnames)
}

func (r *Router) insert(method, path string, dest Destination, t kind, ppath string, pnames []string) {
	// Adjust max param
	l := len(pnames)
	if r.maxParamNum < l {
		r.maxParamNum = l
	}

	cn := r.tree // Current node as root
	search := path

	for {
		sl := len(search)
		pl := len(cn.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == cn.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			cn.label = search[0]
			cn.prefix = search
			if dest != nil {
				cn.kind = t
				cn.addDestination(method, dest)
				cn.ppath = ppath
				cn.pnames = pnames
			}
		} else if l < pl {
			// Split node
			n := newNode(cn.kind, cn.prefix[l:], cn, cn.children, cn.methodDests, cn.ppath, cn.pnames)

			// Reset parent node
			cn.kind = skind
			cn.label = cn.prefix[0]
			cn.prefix = cn.prefix[:l]
			cn.children = nil
			cn.methodDests = new(methodDests)
			cn.ppath = ""
			cn.pnames = nil

			cn.addChild(n)

			if l == sl {
				// At parent node
				cn.kind = t
				cn.addDestination(method, dest)
				cn.ppath = ppath
				cn.pnames = pnames
			} else {
				// Create child node
				n = newNode(t, search[l:], cn, nil, new(methodDests), ppath, pnames)
				n.addDestination(method, dest)
				cn.addChild(n)
			}
		} else if l < sl {
			search = search[l:]
			c := cn.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				cn = c
				continue
			}
			// Create child node
			n := newNode(t, search, cn, nil, new(methodDests), ppath, pnames)
			n.addDestination(method, dest)
			cn.addChild(n)
		} else {
			// Node already exists
			if dest != nil {
				cn.addDestination(method, dest)
				cn.ppath = ppath
				if len(cn.pnames) == 0 {
					cn.pnames = pnames
				}
			}
		}
		return
	}
}

func newNode(k kind, pre string, p *node, c children, md *methodDests, ppath string, pnames []string) *node {
	return &node{
		kind:        k,
		label:       pre[0],
		prefix:      pre,
		parent:      p,
		children:    c,
		ppath:       ppath,
		pnames:      pnames,
		methodDests: md,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, k kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == k {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(k kind) *node {
	for _, c := range n.children {
		if c.kind == k {
			return c
		}
	}
	return nil
}

func (n *node) addDestination(method string, dest Destination) {
	switch method {
	case http.MethodConnect:
		n.methodDests.connect = dest
	case http.MethodDelete:
		n.methodDests.delete = dest
	case http.MethodGet:
		n.methodDests.get = dest
	case http.MethodHead:
		n.methodDests.head = dest
	case http.MethodOptions:
		n.methodDests.options = dest
	case http.MethodPatch:
		n.methodDests.patch = dest
	case http.MethodPost:
		n.methodDests.post = dest
	case MethodPropfind:
		n.methodDests.propfind = dest
	case http.MethodPut:
		n.methodDests.put = dest
	case http.MethodTrace:
		n.methodDests.trace = dest
	}
}

func (n *node) findDestination(method string) Destination {
	switch method {
	case http.MethodConnect:
		return n.methodDests.connect
	case http.MethodDelete:
		return n.methodDests.delete
	case http.MethodGet:
		return n.methodDests.get
	case http.MethodHead:
		return n.methodDests.head
	case http.MethodOptions:
		return n.methodDests.options
	case http.MethodPatch:
		return n.methodDests.patch
	case http.MethodPost:
		return n.methodDests.post
	case MethodPropfind:
		return n.methodDests.propfind
	case http.MethodPut:
		return n.methodDests.put
	case http.MethodTrace:
		return n.methodDests.trace
	default:
		return nil
	}
}

func (r *Router) Find(method, path string, result *Result) bool {
	cn := r.tree // Current node as root
	var (
		search = path
		child  *node  // Child node
		n      int    // Param counter
		nk     kind   // Next kind
		nn     *node  // Next node
		ns     string // Next search
	)
	pvalues := result.PathValues

	// Search order static > param > any
	for {
		if search == "" {
			goto End
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if cn.label != ':' {
			sl := len(search)
			pl = len(cn.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == cn.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			cn = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
			// Not found
			return false
		}

		if search == "" {
			goto End
		}

		// Static node
		if child = cn.findChild(search[0], skind); child != nil {
			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' {
				nk = pkind
				nn = cn
				ns = search
			}
			cn = child
			continue
		}

		// Param node
	Param:
		if child = cn.findChildByKind(pkind); child != nil {
			if len(pvalues) == n {
				continue
			}

			// Save next
			if cn.prefix[len(cn.prefix)-1] == '/' {
				nk = akind
				nn = cn
				ns = search
			}

			cn = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/' && search[i] != '.'; i++ {
			}
			pvalues[n] = search[:i]
			n++
			search = search[i:]
			continue
		}

		// Any node
	Any:
		if cn = cn.findChildByKind(akind); cn == nil {
			if nn != nil {
				cn = nn
				nn = cn.parent
				search = ns
				if nk == pkind {
					goto Param
				} else if nk == akind {
					goto Any
				}
			}
			// Not found
			return false
		}
		pvalues[len(cn.pnames)-1] = search
		goto End
	}

End:
	result.Pattern = cn.ppath
	result.PathParams = cn.pnames
	// result.PathValues = pvalues[:len(cn.pnames)]
	result.Dest = cn.findDestination(method)
	result.MethodNotAllow = false
	// NOTE: Slow zone...
	if result.Dest == nil {
		result.MethodNotAllow = true
		// Dig further for any, might have an empty value for *, e.g.
		// serving a directory.
		if cn = cn.findChildByKind(akind); cn == nil {
			return true
		}
		if dest := cn.findDestination(method); dest != nil {
			result.Dest = dest
			result.MethodNotAllow = false
		}
		pvalues[len(cn.pnames)-1] = ""
		result.Pattern = cn.ppath
		result.PathParams = cn.pnames
		// result.PathValues = pvalues[:len(cn.pnames)]
	}
	return true
}

func (r *Router) NewResult() *Result {
	return &Result{
		PathValues: make([]string, r.maxParamNum, r.maxParamNum),
	}
}

func (r *Result) Param(name string) string {
	for i, n := range r.PathParams {
		if n == name {
			return r.PathValues[i]
		}
	}
	return ""
}

func (r *Result) Reset() {
	r.PathParams = nil
	r.MethodNotAllow = false
	r.Dest = nil
	r.Pattern = ""
	for i := range r.PathValues {
		r.PathValues[i] = ""
	}
}

func (r *Result) ResetValues() {
	for i := range r.PathValues {
		r.PathValues[i] = ""
	}
}
