#!/bin/bash

# 接收一个参数, 去掉其首尾空格
# 返回内容为 trim 后的字符串
# 不会失败
function trimStr() {
    local toTrimStr="$1";

    # 利用 sed 替换
    toTrimStr="$(echo "${toTrimStr}"| sed 's/^\s*//g' | sed 's/\s*$//g')"
    echo "$toTrimStr"

    return 0;
}


# 解析命令行参数$@
# 解析 --xxx yyy 为 xxx=yyy
# 解析 --boolxxx 为 xxx=1
# 不区分大小写,统一改为小写
# 参数名称不能包含空格,否则报错
# 要判断命令结果
function parseParamToVar() {
    local tmpOpt;
    while [ "X$1" != "X" ];do
        # 改为小写
        tmpOpt="$(echo "$1"|tr A-Z a-z)"

        # 判断参数是否是 --bool开头,如果是则将变量置为1 并 shift 1
        if ([[ "$tmpOpt" = --bool* ]]);then
            tmpOpt=${tmpOpt#--}
            # 去掉首尾空格
            tmpOpt="$(trimStr "$tmpOpt")"

            echo "检测到参数 ${tmpOpt} = 1"
            export "${tmpOpt}"="1"
            shift 1
            continue
        fi

        # 判断参数是否是 --开头,如果是则将变量置为下一个参数 并 shift 2
        if ([[ "$tmpOpt" = --* ]]);then
            tmpOpt=${tmpOpt#--}
            # 去掉首尾空格
            tmpOpt="$(trimStr "$tmpOpt")"

            echo "检测到参数 ${tmpOpt} = $2"
            export "${tmpOpt}"="$2"
            shift 2  || shift 1
            # 如果 shift 2 失败,说明没有多余的参数了, 改用 shift 1
            continue
        fi

        shift 1
    done

    return 0;
}

# 检测入参
function checkParams() {
    # 检测 grpc_addr
    if ([ "X$grpc_addr" = "X" ]);then
        echo "未提供 grpc_addr,使用默认值 :9898"
        grpc_addr=":9898"
    fi
    # 检测 grpc_policy_path
    if ([ "X$grpc_policy_path" = "X" ]);then
        echo "未提供 grpc_policy_path,脚本退出"
        exit 1
    fi
    # 检测 addr_port
    if ([ "X$addr_port" = "X" ]);then
        echo "未提供 addr_port,使用默认值 9899"
        addr_port="9899"
    fi
    # 检测 diagnostic_addr_port
    if ([ "X$diagnostic_addr_port" = "X" ]);then
        echo "未提供 diagnostic_addr_port,使用默认值 9900"
        diagnostic_addr_port="9900"
    fi
    # 检测 config
    if ([ "X$config" = "X" ]);then
        echo "未提供 config,脚本退出"
        exit 1
    fi
}

# #################
# 脚本开始
parseParamToVar $@

# 检测参数
checkParams

# 创建文件
if ([ "X$files" = "X" ]);then
    echo "未提供 files,不执行文件创建操作"
else
    # base64解码获得files的真实内容
    files=$(echo "${files}" | base64 -d)

    # 获取json所有的keys jsonArr
    jsonArrForAllKeys=$(echo $files | jq 'keys')
    jsonArrForAllKeysLen=$(echo $jsonArrForAllKeys | jq 'length')

    # 检查解析结果,防止 files 不是有效的json格式
    if ([ $? -ne 0 ]);then
        echo "files参数内容不是有效的json格式"
        exit 1
    fi

    # 遍历json
    for (( index=0 ; index<${jsonArrForAllKeysLen}; index++ )); do
        # 获取jsonKey
        tmpJsonKey=$(echo $jsonArrForAllKeys | jq ".[$index]" -r)
        # 获取key对应的值
        tmpValueForKey="$(echo $files | jq ".$tmpJsonKey" -cr)"
        # 将值base64解码后输出到文件
        echo "${tmpValueForKey}" | base64 -d > "${tmpJsonKey}"
    done
fi


# 生成配置文件
# base64解码获得config的真实内容
config=$(echo "${config}" | base64 -d)
echo "${config}"  > /app/user-config-file.json

# 启动 opa
if ([ "X${POD_NAME}}" = "X" ]);then
    /app/opa run --server --config-file=/app/user-config-file.json \
    --set plugins.envoy_ext_authz_grpc.addr=${grpc_addr} \
    --set plugins.envoy_ext_authz_grpc.path=${grpc_policy_path} \
    --set decision_logs.console=true \
    --addr=localhost:${addr_port} \
    --diagnostic-addr=0.0.0.0:${diagnostic_addr_port}
else
    /app/opa run --server --config-file=/app/user-config-file.json \
    --set plugins.envoy_ext_authz_grpc.addr=${grpc_addr} \
    --set plugins.envoy_ext_authz_grpc.path=${grpc_policy_path} \
    --set decision_logs.console=true \
    --addr=localhost:${addr_port} \
    --diagnostic-addr=0.0.0.0:${diagnostic_addr_port} \
    --set labels.pod_name=${POD_NAME} \
    --set labels.pod_namespace=${POD_NAMESPACE}
fi