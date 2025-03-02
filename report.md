# IDBS-Spring20-Fudan Assignment3报告

姓名：夏海淞

学号：18307130090

代码地址：https://github.com/xhs7700/FudanLMS

## 关系模型设计

### ER图

![ER Graph](C:\Users\夏海淞\Documents\My Markdown\课堂笔记\数据库引论\图片\ER Graph.jpg)

### 关系模式集

users(<u>id</u>, isbn, authority)

books(<u>isbn</u>, title, author)

borrec(<u><span style="text-decoration-line: underline; text-decoration-style: wavy;">id</span>, <span style="text-decoration-line: underline; text-decoration-style: wavy;">isbn</span></u>, bortime, deadline, extendtime)

retrec(<u><span style="text-decoration-line: underline; text-decoration-style: wavy;">id</span>, <span style="text-decoration-line: underline; text-decoration-style: wavy;">isbn</span>, bortime</u>, rettime)

## 系统功能与实现方法

### 登录

用户可以使用管理员注册的账号进行登录。只有登录后才能进行借书、还书等操作。

#### 实现方法

用户在终端中输入用户名和密码后，后端在用户表中查询是否存在该用户。如果不存在该用户则报错。

否则，将数据库中存储的密码哈希值与输入的密码哈希值进行比对，如果不一致，则报错；否则更新当前登录用户状态，提示登陆成功。

### 注册

管理员可以注册新的管理员账号和新的学生账号。

#### 实现方法

管理员在终端中通过"-[as]"的参数控制新用户的权限。在管理员输入用户名、密码和重复密码后，后端首先在用户表中查询是否存在相同的用户名，如果存在则报错。

否则，比对管理员两次输入的密码是否一致，如果不一致则报错。否则将用户名、密码的哈希值和权限代码插入数据库的用户表中，提示注册成功。

### 修改密码

任何用户均可修改自己的密码。在修改密码时需要提供当前用户的现有密码以验证身份。

#### 实现方法

在用户输入旧密码和新密码后，后端首先在用户表中查询当前用户的密码哈希值，并与输入的旧密码比对。比对完成后，后端将用户表中密码字段修改为新密码的哈希值。

### 重置密码

管理员可以重置任何用户的密码。这样普通用户在忘记密码时可与管理员联系，让管理员验证身份后重置密码。

管理员需要提供重置密码的用户名和重置后的密码。

#### 实现方法

后端首先在用户表中确认是否存在该用户，如果不存在则报错。否则将该用户元组中密码字段更改为新密码的哈希值。

### 查询书籍

任何用户（包括游客）均可通过书名、作者、ISBN查询数据库中的书籍。

用户通过"-[ait]"形式的参数控制后端分别依据作者、ISBN、书名查询书籍。注意此处查询仅支持完整匹配。若查询值为"*"则表示查询数据库中存储的所有书籍。

#### 实现方法

后端获取参数和查询值后，在数据库中查询，将返回值呈现给前端。

### 查询借阅记录

普通用户可以查询自己的借阅记录，管理员可以查询任何用户的借阅记录。

用户通过"-[bar]"形式的参数控制后端分别返回尚在借阅状态的借阅记录、已归还的借阅记录和所有借阅记录。

#### 实现方法

后端获取参数和用户名后，首先查找用户表，确认是否存在该用户；随后在借阅记录表或归还记录表中查找该用户的记录，以列表的形式返回前端。

### 借阅书籍

用户可以通过输入ISBN编码借阅对应的书籍。还书期限被设定为借书日期的30天后。用户不可以同时借阅两本相同的书籍。已经因超期未还书籍过多而被限制借阅功能的用户不能借阅书籍。

#### 实现方法

后端获取用户名和ISBN编码后，首先确认是否存在该书籍以及是否已存在相同的借阅记录；随后在借阅记录表中插入新元组，提示前端借阅成功。

### 归还书籍

普通用户可以通过输入ISBN编码归还对应的书籍。管理员用户可以通过输入用户名和ISBN编码归还任何用户借阅的对应书籍。

#### 实现方法

后端获取用户名和ISBN编码后，首先在借阅记录表中确认是否存在该借阅记录，如果不存在则报错；随后在借阅记录表中删除相应的借阅记录，在归还记录表中插入新元组，并将归还时间设为系统当前时间。

### 添加书籍

管理员可以通过输入ISBN编码、书名和作者添加新的书籍。

#### 实现方法

后端获取ISBN编码、书名和作者名后，首先在书籍表中确认是否已存在该书，如果存在则报错；随后将这些信息作为一个元组插入书籍表中，提示前端添加成功。

### 删除书籍

管理员可以通过输入ISBN编码、删除原因删除现有的书籍。

#### 实现方法

后端获取ISBN编码和删除原因，首先在书籍表中确认是否存在该书，如果不存在则报错；随后在书籍表中删除该元组。

### 查询还书期限

普通用户可以通过输入ISBN编码的方式查询该书的还书期限；管理员可以通过输入用户名和ISBN编码的方式查询任何用户对应借阅书籍的还书期限。

#### 实现方法

后端获取用户名（普通用户默认为自己的用户名）和ISBN编码后，首先在借阅记录表中确认是否存在该借阅记录，如果不存在则报错；随后返回该借阅记录中的还书期限字段。

### 查询超期书籍

普通用户可以查询自己的超期书籍；管理员可以通过输入用户名的方式查询任何用户的超期书籍。

#### 实现方法

后端获取用户名（普通用户默认为自己的用户名）后，首先在用户表中确认是否存在该用户，如果不存在则报错；随后在借阅记录表中查询该用户的所有最后期限小于当前时间的借阅记录，以列表的形式返回给前端。

### 延长还书期限

普通用户每次可以将将自己借阅记录的还书期限延长一周。该延长计入该借阅记录的总延长次数。

管理员每次可以将任意借阅记录的还书期限延长任意周数。该延长不计入该借阅记录的总延长次数。

任意借阅记录的总延长次数不得超过3次。

#### 实现方法

后端获取借阅记录对应的用户名和ISBN编码、操作人的权限和延长周数（普通用户默认为一周）后，首先在借阅记录表中确认是否存在该借阅记录，如果不存在则报错。

随后根据操作人的权限，决定延长的时间以及该延长是否计入总延长次数。后端将借阅记录表中相应元组更新后，后端返回更新后的借阅记录。

## 系统优点和不足

### 优点

1. 充分考虑了各种异常情况，通过Go语言的错误机制反馈给用户和开发者，方便后续调试。
2. 尽量减少命令参数的使用，通过输入输出交互的方式和用户交互，降低用户的学习成本。
3. 提供帮助文本，更加用户友好。
4. 通过第三方库实现了隐藏用户输入密码内容，确保用户隐私。

### 不足

1. 缺少批处理操作，管理员进行大规模添加删除书籍时极为不便。后续考虑添加批处理的功能。
2. 查找书籍时只能通过完全匹配的方式查找，对用户而言并不友好。后续考虑添加搜索引擎的功能。