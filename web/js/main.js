var T = window.T;
var dt_interval;
var map;
var local_search;
var geocoder;
var markers;
var socket;
var logname;
var tmpljson, dictjson;
var re = new RegExp("^(([1-9]\\d*)|0)\\.\\d+/[+-](([1-9]\\d*)|0)\\.\\d+$");
var nextlog = 1;
var marks = {};

var dynamics = { "rrig": new Set(), "rpwr": new Set(), "rant": new Set(), "rqth": new Set() };

const ALLOWED_MODES = [
    "AM", "ARDOP", "ATV", "CHIP", "CLO", "CONTESTI", "CW", "DIGITALVOICE", "DOMINO", "DYNAMIC", "FAX",
    "FM", "FSK441", "FSK", "FT8", "HELL", "ISCAT", "JT4", "JT6M", "JT9", "JT44", "JT65", "MFSK", "MSK144",
    "MTONE", "MT63", "OLIVIA", "OPERA", "PAC", "PAX", "PKT", "PSK", "PSK2K", "Q15", "QRA64", "ROS", "RTTY",
    "RTTYM", "SSB", "SSTV", "T10", "THOR", "THRB", "TOR", "V4", "VOI", "WINMOR", "WSPR", "AMTORFEC", "ASCI",
    "C4FM", "CHIP64", "CHIP128", "DOMINOF", "DSTAR", "FMHELL", "FSK31", "GTOR", "HELL80", "HFSK", "JT4A",
    "JT4B", "JT4C", "JT4D", "JT4E", "JT4F", "JT4G", "JT65A", "JT65B", "JT65C", "MFSK8", "MFSK16", "PAC2",
    "PAC3", "PAX2", "PCW", "PSK10", "PSK31", "PSK63", "PSK63F", "PSK125", "PSKAM10", "PSKAM31", "PSKAM50",
    "PSKFEC31", "PSKHELL", "QPSK31", "QPSK63", "QPSK125", "THRBX"
];

// 时间更新函数
function update_dt() {
    const dt = document.getElementById("dt");
    const now = new Date();
    dt.value = now.toISOString().substring(0, 16);
}

// 输入框清理函数
function clear_input() {
    document.querySelectorAll("input").forEach((ele) => { if (ele.type == "text" && !ele.disabled) ele.value = ""; });
    document.getElementById("index").value = 0; // by default, means new log line
    document.getElementById("rst").value = 59; // by default
}

// 自动更新时间
function auto_dt(cb) {
    const ele_dt = document.getElementById("dt");
    if (cb.checked == true) {
        update_dt();
        ele_dt.disabled = true;
        dt_interval = setInterval(update_dt, 10000);
    } else {
        ele_dt.disabled = false;
        clearInterval(dt_interval);
    }
}

// 锁定参数
function click_lock(cb, id) {
    document.getElementById(id).disabled = cb.checked;
}

// 编辑条目
function editlog(idx) {
    let keys = ["index", "callsign", "dt", "freq", "mode", "rst", "rrig", "rpwr", "rant", "rqth", "trig", "tpwr", "tant", "tqth", "rmks"];
    for (let i = 0; i < keys.length; i++) {
        document.getElementById(`${keys[i]}`).value = document.getElementById(`log_td_i${idx}_${keys[i]}`).innerText;
    }
    document.getElementById("inputtitle").innerText = `输入 #${idx}`;
    document.getElementById("submit").value = `提交 #${idx}`;
}

// 标记删除条目
function deletelog(idx) {
    let retjson = {
        "type": 3,
        "message": "",
        "payload": {
            "index": -parseInt(document.getElementById(`log_td_i${idx}_index`).innerText),
            "callsign": document.getElementById(`log_td_i${idx}_callsign`).innerText,
            "dt": document.getElementById(`log_td_i${idx}_dt`).innerText,
            "freq": document.getElementById(`log_td_i${idx}_freq`).innerText,
            "mode": document.getElementById(`log_td_i${idx}_mode`).innerText,
            "rst": parseInt(document.getElementById(`log_td_i${idx}_rst`).innerText),
            "rrig": document.getElementById(`log_td_i${idx}_rrig`).innerText,
            "rpwr": document.getElementById(`log_td_i${idx}_rpwr`).innerText,
            "rant": document.getElementById(`log_td_i${idx}_rant`).innerText,
            "rqth": document.getElementById(`log_td_i${idx}_rqth`).innerText,
            "trig": document.getElementById(`log_td_i${idx}_trig`).innerText,
            "tpwr": document.getElementById(`log_td_i${idx}_tpwr`).innerText,
            "tant": document.getElementById(`log_td_i${idx}_tant`).innerText,
            "tqth": document.getElementById(`log_td_i${idx}_tqth`).innerText,
            "rmks": document.getElementById(`log_td_i${idx}_rmks`).innerText
        }
    }
    if (!ALLOWED_MODES.includes(retjson.payload.mode)) {
        alert("未知的模式");
        return;
    }
    if (!re.test(retjson.payload.freq)) {
        alert("无法解析的频率");
        return;
    }
    socket.send(JSON.stringify(retjson));
}

// 添加到模板
function add2tmpl(idx) {
    let newtmpl = {
        "callsign": document.getElementById(`log_td_i${idx}_callsign`).innerText,
        "rig": document.getElementById(`log_td_i${idx}_rrig`).innerText,
        "pwr": document.getElementById(`log_td_i${idx}_rpwr`).innerText,
        "ant": document.getElementById(`log_td_i${idx}_rant`).innerText,
        "qth": document.getElementById(`log_td_i${idx}_rqth`).innerText
    };
    if (!tmpljson.some(item => item.callsign == newtmpl.callsign && item.rig == newtmpl.rig && item.pwr == newtmpl.pwr && item.ant == newtmpl.ant && item.qth == newtmpl.qth)) {
        tmpljson.push(newtmpl);
        fetch(`http://${window.location.host}/editdb`, {
            method: "POST",
            headers: {
                "Content-type": "application/x-www-form-urlencoded",
            },
            body: `type=tmpl&payload=${encodeURIComponent(JSON.stringify(tmpljson))}`
        });
    }
}

function autocomplete_cb(name, str) {
    return () => {
        let rst = [];
        let rstkeys = Object.keys(dictjson[name]).filter((item) => { return item.includes(document.getElementById("r" + name).value.toLowerCase()); });
        for (let i of rstkeys) {
            rst = rst.concat(dictjson[name][i]);
        }
        if (rst) {
            let suggests_ol = document.getElementById("suggests");
            suggests_ol.innerHTML = "";
            document.getElementById("mapsuggests").innerHTML = "";
            for (let i of rst) {
                suggests_ol.insertAdjacentHTML("beforeend", `<li>[${str}]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('r${name}').value='${i}';document.getElementById('r${name}').focus();">${i}</a></li>`);
            }
        }
    };
}

function dynamichistory_cb(name) {
    return () => {
        document.getElementById("suggests").innerHTML = "";
        let dynamics_ol = document.getElementById("dynamics");
        dynamics_ol.innerHTML = "";
        for (const item of dynamics[name]) {
            dynamics_ol.insertAdjacentHTML("beforeend", `<li>[动态]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('${name}').value='${item}';document.getElementById('${name}').focus();">${item}</a></li>`);
        }
    };
}

function quickin_cb(name) {
    return (event) => {
        if ((event.ctrlKey || event.metaKey) && ['1', '2', '3', '4', '5', '6', '7', '8', '9'].includes(event.key)) {
            event.preventDefault();
            let idx = parseInt(event.key) - 1;
            let suggests = document.getElementById("suggests");
            if (suggests.childElementCount == 0) suggests = document.getElementById("dynamics");
            if (idx < suggests.childElementCount) document.getElementById(name).value = suggests.children[idx].children[0].innerText;
        }
    };
}

// 窗口加载完成执行
function onload() {
    // 获取记忆数据
    fetch(`http://${window.location.host}/tmpl.json`)
        .then((response) => {
            if (response.ok) return response.json();
            else return null;
        })
        .then((data) => {
            tmpljson = data;
            document.getElementById("tmpldone").innerText = " | 模板已加载";
        });
    fetch(`http://${window.location.host}/dict.json`)
        .then((response) => {
            if (response.ok) return response.json();
            else return null;
        })
        .then((data) => {
            dictjson = data;
            document.getElementById("dictdone").innerText = " | 字典已加载";
        });

    // 天地图
    if (T) {
        map = new T.Map("mapdiv");
        map.centerAndZoom(new T.LngLat(116.40769, 39.89945), 12);
        geocoder = new T.Geocoder();
        markers = new T.MarkerClusterer(map, { markers: [] });

        // 设置滚动
        document.getElementById("mapdiv").addEventListener("wheel", (e) => {
            let direction = e.deltaY > 0 ? "up" : "down";
            direction === "up" ? map.zoomOut() : map.zoomIn();
            e.preventDefault();
        });

        // 设置控件
        let ctrl_zoom = new T.Control.Zoom();
        ctrl_zoom.setPosition(T_ANCHOR_BOTTOM_RIGHT);
        map.addControl(ctrl_zoom);
        let ctrl_scale = new T.Control.Scale();
        map.addControl(ctrl_scale);
        let ctrl_maptype = new T.Control.MapType();
        map.addControl(ctrl_maptype);

        // 搜索
        local_search = new T.LocalSearch(map, {
            pageCapacity: 10, onSearchComplete: (rst) => {
                let rst_suggests = rst.getSuggests();
                if (rst_suggests) {
                    let mapsuggests_ol = document.getElementById("mapsuggests");
                    mapsuggests_ol.innerHTML = "";
                    for (let i = 0; i < rst_suggests.length; i++) {
                        mapsuggests_ol.insertAdjacentHTML("beforeend", `<li>[天地图]&nbsp;<a href="javascript:void(0);" onclick="let element=document.getElementById('rqth');element.value='${rst_suggests[i].address + rst_suggests[i].name}';element.focus();">${rst_suggests[i].name}</a><i>${rst_suggests[i].address}</i></li>`);
                    }
                }
            }
        });
        document.getElementById("rqth").addEventListener("input", () => {
            local_search.search(document.getElementById("rqth").value, 4);
        });
    }

    // 简写搜索
    // 模式
    document.getElementById("mode").addEventListener("input", () => {
        let ele = document.getElementById("mode");
        ele.value = ele.value.toUpperCase();
        let rst = ALLOWED_MODES.filter((item) => { return item.includes(ele.value); });
        let suggests_ol = document.getElementById("suggests");
        suggests_ol.innerHTML = "";
        document.getElementById("dynamics").innerHTML = "";
        document.getElementById("mapsuggests").innerHTML = "";
        for (let i of rst) {
            suggests_ol.insertAdjacentHTML("beforeend", `<li>[模式]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('mode').value='${i}';document.getElementById('mode').focus();">${i}</a></li>`);
        }
    });
    // 呼号 - 模板
    document.getElementById("callsign").addEventListener("input", () => {
        let ele = document.getElementById("callsign");
        ele.value = ele.value.toUpperCase();
        let rst = tmpljson.filter((item) => { return item.callsign.includes(ele.value); });
        let suggests_ol = document.getElementById("suggests");
        suggests_ol.innerHTML = "";
        document.getElementById("dynamics").innerHTML = "";
        document.getElementById("mapsuggests").innerHTML = "";
        for (let i of rst) {
            suggests_ol.insertAdjacentHTML("beforeend", `<li>[模板]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('callsign').value='${i.callsign}';document.getElementById('rrig').value='${i.rig}';document.getElementById('rpwr').value='${i.pwr}';document.getElementById('rant').value='${i.ant}';document.getElementById('rqth').value='${i.qth}';document.getElementById('callsign').focus();">${i.callsign}</a><i>${i.rig}|${i.pwr}|${i.ant}|${i.qth}</i></li>`);
        }
    });
    // 字典自动补全
    document.getElementById("rrig").addEventListener("input", autocomplete_cb("rig", "设备"));
    document.getElementById("rpwr").addEventListener("input", autocomplete_cb("pwr", "功率"));
    document.getElementById("rant").addEventListener("input", autocomplete_cb("ant", "天线"));
    document.getElementById("rqth").addEventListener("input", autocomplete_cb("qth", "QTH"));

    // 自动历史输入
    document.getElementById("rrig").addEventListener("focus", dynamichistory_cb("rrig"));
    document.getElementById("rpwr").addEventListener("focus", dynamichistory_cb("rpwr"));
    document.getElementById("rant").addEventListener("focus", dynamichistory_cb("rant"));
    document.getElementById("rqth").addEventListener("focus", dynamichistory_cb("rqth"));

    // 输入选择快捷键
    document.getElementById("rrig").addEventListener("keydown", quickin_cb("rrig"));
    document.getElementById("rpwr").addEventListener("keydown", quickin_cb("rpwr"));
    document.getElementById("rant").addEventListener("keydown", quickin_cb("rant"));
    document.getElementById("rqth").addEventListener("keydown", quickin_cb("rqth"));

    // 检查复选框状态
    document.getElementById("dt").disabled = document.getElementById("dtauto").checked;
    let lock_names = ["freq", "mode", "trig", "tpwr", "tant", "tqth"];
    for (let i of lock_names) document.getElementById(i).disabled = document.getElementById(i + "lock").checked;

    // 清空输入框
    clear_input();
    // 时间同步更新
    update_dt();
    auto_dt(document.getElementById("dtauto"));

    // 表单提交行为
    document.getElementById("infoform").addEventListener("submit", (ev) => {
        ev.preventDefault();
        let retjson = {
            "type": 3,
            "message": "",
            "payload": {
                "index": parseInt(document.getElementById("index").value),
                "callsign": document.getElementById("callsign").value,
                "dt": document.getElementById("dt").value,
                "freq": document.getElementById("freq").value,
                "mode": document.getElementById("mode").value,
                "rst": parseInt(document.getElementById("rst").value),
                "rrig": document.getElementById("rrig").value,
                "rpwr": document.getElementById("rpwr").value,
                "rant": document.getElementById("rant").value,
                "rqth": document.getElementById("rqth").value,
                "trig": document.getElementById("trig").value,
                "tpwr": document.getElementById("tpwr").value,
                "tant": document.getElementById("tant").value,
                "tqth": document.getElementById("tqth").value,
                "rmks": document.getElementById("rmks").value
            }
        }
        if (!ALLOWED_MODES.includes(retjson.payload.mode)) {
            alert("未知的模式");
            return;
        }
        if (!re.test(retjson.payload.freq)) {
            alert("无法解析的频率");
            return;
        }
        dynamics["rrig"].add(document.getElementById("rrig").value);
        dynamics["rpwr"].add(document.getElementById("rpwr").value);
        dynamics["rant"].add(document.getElementById("rant").value);
        dynamics["rqth"].add(document.getElementById("rqth").value);
        socket.send(JSON.stringify(retjson));
        clear_input();
        document.getElementById("callsign").focus();
    });

    // 询问文件名
    logname = prompt("输入要打开的文件名（不存在则新建）", "newlogbook");
    if (logname == null || logname == "") {
        alert("您没有输入文件名");
        return;
    }
    document.getElementById("exportlink").href = `/export?n=${logname}`;

    // ws
    socket = new WebSocket(`ws://${window.location.host}/ws`);
    socket.onopen = () => { socket.send(JSON.stringify({ "type": 1, "message": logname })); };
    socket.onmessage = (ev) => {
        let info = JSON.parse(String(ev.data));
        if (info.type == 2 && info.message == "OK") {
            document.getElementById("connected").innerText = ` | 已打开 ${logname}.hjl`;
        } else if (info.type == 3) {
            let markdel = false;
            // 更新提交提示
            nextlog = nextlog > info.payload.index ? nextlog : info.payload.index + 1;
            document.getElementById("inputtitle").innerText = `输入 #${nextlog}`;
            document.getElementById("submit").value = `提交 #${nextlog}`;
            // 记录表格
            let keys = ["callsign", "dt", "freq", "mode", "rst", "rrig", "rpwr", "rant", "rqth", "trig", "tpwr", "tant", "tqth", "rmks"];
            let inner = `<td><input type="button" value="标记为删除" onclick="deletelog(${info.payload.index})"><input type="button" value="添加到模板" onclick="add2tmpl(${info.payload.index})"></td><td id="log_td_i${info.payload.index}_index"><a href="javascript:void(0);" onclick="editlog(${info.payload.index})">${info.payload.index}</a></td>`;
            for (let i = 0; i < keys.length; i++) {
                inner += `<td id="log_td_i${info.payload.index}_${keys[i]}">${info.payload[keys[i]]}</td>`;
            }
            if (info.message == "SYNC" || info.message == "ADD") {
                let newtr = document.createElement("tr");
                newtr.id = `log_tr_i${info.payload.index}`;
                newtr.innerHTML = inner;
                document.getElementById("logtable").appendChild(newtr);
            } else if (info.message == "EDIT") {
                if (info.payload.index < 0) {
                    info.payload.index = -info.payload.index;
                    markdel = true;
                    inner = `<td></td><td id="log_td_i${info.payload.index}_index"><a href="javascript:void(0);" onclick="editlog(${info.payload.index})"><del>${info.payload.index}</del></a></td>`;
                    for (let i = 0; i < keys.length; i++) {
                        inner += `<td id="log_td_i${info.payload.index}_${keys[i]}"><del>${info.payload[keys[i]]}<del></td>`;
                    }
                }
                document.getElementById(`log_tr_i${info.payload.index}`).innerHTML = inner;
            }
            // 自动滚动
            let logdiv = document.getElementById("logdiv");
            logdiv.scroll({
                top: logdiv.scrollHeight,
                left: 0,
                behavior: "smooth",
            });
            if (T && (info.message == "ADD" || info.message == "EDIT") && !markdel) {
                // 地理编码
                geocoder.getPoint(info.payload.rqth, (rst) => {
                    if (rst.getStatus() == 0) {
                        if (info.payload.index in marks) {
                            markers.removeMarker(marks[info.payload.index]);
                        }
                        map.panTo(rst.getLocationPoint(), 16);
                        marks[info.payload.index] = new T.Marker(rst.getLocationPoint());
                        let marker_info = new T.InfoWindow(`${info.payload.index}. ${info.payload.callsign}<br>${info.payload.dt}<br>${info.payload.rqth}`);
                        marks[info.payload.index].addEventListener("click", () => { marks[info.payload.index].openInfoWindow(marker_info); });
                        markers.addMarker(marks[info.payload.index]);
                    }
                });
            }
        }
    };
    socket.onerror = () => { alert("服务连接失败"); };
}
