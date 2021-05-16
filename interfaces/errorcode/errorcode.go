package errorcode

// 系统类型的错误码
const CODE_SUCCESS = 0
const CODE_PARAMS_INVALID = -1
const CODE_AUTH_CHECK_TOKEN_FAIL = -2
const CODE_API_TOKEN_FAIL = -3

// 用户类型的错误
const CODE_ADMIN_ROLE_LOW = 996        // 权限过低
const CODE_ADMIN_MOBILE_EXIST = 995    // 手机号已存在
const CODE_ADMIN_PASSWORD_FORMAT = 997 // 密码格式错误

// 百度贴吧错误
const CODE_BAIDUTIEBA_ARTICLE_EXIST = 1010 // 帖子已存在
