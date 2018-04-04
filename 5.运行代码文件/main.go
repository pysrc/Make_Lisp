package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

var DEBUG = false

func Debug(v ...interface{}) {
	if DEBUG {
		fmt.Println(v...)
	}
}

type Object interface{}
type List struct { // 列表，不参与运算的列表[1 2 3]
	Val []Object
}
type Fn struct { // 函数结构
	Name string             // 函数名
	Args *List              // 形参
	Body Object             // 函数体
	Env  *map[string]Object // 函数环境
}

// 全局变量/函数
var EnvMap = map[string]Object{
	"+": func(v []Object) Object {
		return v[0].(float64) + v[1].(float64)
	},
	"-": func(v []Object) Object {
		return v[0].(float64) - v[1].(float64)
	},
	"*": func(v []Object) Object {
		return v[0].(float64) * v[1].(float64)
	},
	"/": func(v []Object) Object {
		return v[0].(float64) / v[1].(float64)
	},
	"^": func(v []Object) Object { // 指数
		return math.Pow(v[0].(float64), v[1].(float64))
	},
	">": func(v []Object) Object {
		return v[0].(float64) > v[1].(float64)
	},
	">=": func(v []Object) Object {
		return v[0].(float64) >= v[1].(float64)
	},

	"<": func(v []Object) Object {
		return v[0].(float64) < v[1].(float64)
	},
	"<=": func(v []Object) Object {
		return v[0].(float64) <= v[1].(float64)
	},
	"==": func(v []Object) Object {
		return v[0].(float64) == v[1].(float64)
	},
	"!=": func(v []Object) Object {
		return v[0].(float64) != v[1].(float64)
	},
	"sin": func(v []Object) Object {
		return math.Sin(v[0].(float64))
	},
	"cos": func(v []Object) Object {
		return math.Cos(v[0].(float64))
	},
	"tan": func(v []Object) Object {
		return math.Tan(v[0].(float64))
	},
	"out": func(v []Object) Object { // 输出函数
		return v
	},
	// 其余可自行添加
}

func IsSym(v byte) bool { // 括号判断
	syms := "(){}[]"
	for i := 0; i < len(syms); i++ {
		if syms[i] == v {
			return true
		}
	}
	return false
}

func IsNum(v byte) bool { // 数字判断
	if v >= '0' && v <= '9' {
		return true
	}
	return false
}

type Code struct {
	src    string // 源码
	pos    int    // 当前指针位置
	tokens []Object
}

/**词法分析开始**/
func (self *Code) Init(src string) { // 初始化
	self.src = src
	self.pos = 0
	self.Next() // 初始化就读一个token
}
func (self *Code) Peek() Object { // 返回当前token
	if len(self.tokens) > 0 {
		return self.tokens[len(self.tokens)-1]
	}
	return nil
}
func (self *Code) Next() Object { // 下一个token
	if self.pos >= len(self.src) {
		return nil
	}
	var tmp []byte
	var tk Object
	var i = self.pos

	switch {
	case IsNum(self.src[i]) || (self.src[i] == '-' && IsNum(self.src[i+1])): // 读取数字
		for i < len(self.src) && (IsNum(self.src[i]) || self.src[i] == '.' || self.src[i] == '-') {
			tmp = append(tmp, self.src[i])
			i++
		}
		self.pos = i
		tk, _ = strconv.ParseFloat(string(tmp), 64)
	case IsSym(self.src[i]): // 括号判断
		tmp = append(tmp, self.src[i])
		i++
		self.pos = i
		tk = string(tmp)
	case self.src[i] == ' ': // 空格
		self.pos++
		tk = self.Next()
	default: // 变量||函数
		for i < len(self.src) && self.src[i] != ' ' && !IsSym(self.src[i]) {
			tmp = append(tmp, self.src[i])
			i++
		}
		self.pos = i
		tk = string(tmp)
	}
	self.tokens = append(self.tokens, tk)
	return tk
}
func (self *Code) read_list() []Object { // 读列表
	var lt []Object
	v := self.Peek()
	for v != ")" && v != nil {
		if v == "(" {
			self.Next()
			lt = append(lt, self.read_list())
		} else if v == "[" { // 不运算列表
			var tpl List
			v = self.Next()
			for v != "]" && v != nil {
				tpl.Val = append(tpl.Val, v)
				v = self.Next()
			}
			lt = append(lt, tpl)
		} else {
			lt = append(lt, v)
		}
		v = self.Next()
	}
	return lt
}

func (self *Code) Read_Root() Object { // 根节点开始解析
	v := self.Peek()
	if v == "(" { // 运算列表
		self.Next()
		return self.read_list()
	} else {
		return v
	}
}

/**词法分析结束**/
/**计算开始**/
func Apply(v []Object, env *map[string]Object, fn func(Object, *map[string]Object) Object) []Object { // 将函数fn 应用到列表每一项
	var res []Object
	for _, j := range v {
		res = append(res, fn(Eval(j, env), env))
	}
	return res
}
func Eval(tree Object, env *map[string]Object) Object { // 计算表达式
	/**
	  tree token对象
	  env 环境（变量、函数）
	*/
	Debug("AST:", tree)
	switch tree.(type) {
	case []Object:
		v, _ := tree.([]Object)
		// fmt.Println("switch:", len(v))
		if len(v) < 2 {
			return nil
		}
		// 取出操作符及其对应的函数
		op := v[0].(string)
		// fmt.Println("op:", op)
		switch op {
		case "set": // 设置变量值(set a 12)或者(set f (+ 1 2))
			key := v[1].(string)
			val := Eval(v[2], env)
			(*env)[key] = val
			return val
		case "if": // 判断语句(if true 3 5)
			du, ok := Eval(v[1], env).(bool)
			if ok {
				if du {
					return Eval(v[2], env)
				} else {
					return Eval(v[3], env)
				}
			}
		case "fn": // 函数定义 (fn fn_name [x, y] (+ x y))
			fn_env := make(map[string]Object)
			var fn Fn
			for k, v := range *env { // 拷贝外环境
				fn_env[k] = v
			}
			fn.Env = &fn_env
			fn.Name = v[1].(string)
			args := v[2].(List)
			fn.Args = &args
			// for _, j := range args.Val {
			// 	fn_env[j.(string)] = nil
			// }
			fn.Body = v[3]
			(*env)[fn.Name] = fn    // 向上一层环境中加入函数
			(*fn.Env)[fn.Name] = fn // 要想实现递归,就应当在自己的环境中找到自己,这是必须的
			return fn
		default:
			if f, ok := (*env)[op]; ok {
				switch f.(type) {
				case func([]Object) Object: // 系统函数
					fc := f.(func([]Object) Object)
					return fc(Apply(v[1:], env, Eval))
				case Fn: // 自定义函数 (fn_name args1 args2 ...)
					// fmt.Println("自定义函数:", op)
					fc := f.(Fn)
					fenv := make(map[string]Object) // 环境拷贝
					for k, v := range *(fc.Env) {
						fenv[k] = v
					}
					// 取传入函数的参数(可能是表达式)
					args := v[1:]
					for i, j := range args { // 将传递的参数加入函数环境
						fenv[fc.Args.Val[i].(string)] = Eval(j, env)
					}
					return Eval(fc.Body, &fenv)
				}

			}
		}

	case Object:
		switch tree.(type) {
		case string:
			if k, ok := (*env)[tree.(string)]; ok {
				//存在变量
				return k
			}
		}
		return tree
	}
	return nil
}
func FindExpr(s string) string { // 找出字符串中的语句
	var x, y int
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			x = i
			break
		}
	}
	for i := len(s) - 1; i > 0; i-- {
		if s[i] == ')' {
			y = i
			break
		}
	}
	if x < y { // 找到
		return s[x:y]
	} else {
		return "" // 找不到
	}
}
func ExeFile(cmd string) { // 执行文件
	cmds := strings.Split(cmd, "\n") // 分离语句
	for _, j := range cmds {
		pr := FindExpr(j) // 任何非表达式都是注释
		if pr != "" {
			c := Code{}
			c.Init(pr)
			s := c.Read_Root()
			if strings.Contains(pr, "set") || strings.Contains(pr, "fn") {
				Eval(s, &EnvMap)
			} else {
				fmt.Println(Eval(s, &EnvMap))
			}
		}
	}
}

func ExeIDLE() { // 解释执行
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("User>>")
		cmds, _, _ := reader.ReadLine()
		cmd := string(cmds)
		if cmd == "exit" {
			return
		} else {
			cmd = FindExpr(cmd)
			if cmd == "" {
				return
			}
			c := Code{}
			c.Init(cmd)
			s := c.Read_Root()
			fmt.Println(Eval(s, &EnvMap))
		}

	}
}

/*计算结束*/
func main() {
	args := os.Args
	if len(args) < 2 {
		ExeIDLE()
	} else {
		src, err := ioutil.ReadFile(args[1])
		if err == nil {
			ExeFile(string(src))
		} else {
			fmt.Println("打开文件出错！")
		}
	}
}
