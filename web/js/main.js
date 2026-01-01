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
    document.getElementById("index").value = 0; // by default
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
    document.getElementById("submit").value = `提交 #${idx}`;
}

// 窗口加载完成执行
function onload() {
    // 获取记忆数据
    let xhrtmpl = new XMLHttpRequest();
    xhrtmpl.onreadystatechange = () => {
        if (xhrtmpl.readyState == 4 && xhrtmpl.status == 200) {
            tmpljson = JSON.parse(xhrtmpl.responseText);
            document.getElementById("tmpldone").innerText = " | 模板已加载";
        }
    };
    xhrtmpl.open("GET", `http://${window.location.host}/tmpl.json`, true);
    xhrtmpl.send();
    let xhrdict = new XMLHttpRequest();
    xhrdict.onreadystatechange = () => {
        if (xhrdict.readyState == 4 && xhrdict.status == 200) {
            dictjson = JSON.parse(xhrdict.responseText);
            document.getElementById("dictdone").innerText = " | 字典已加载";
        }
    };
    xhrdict.open("GET", `http://${window.location.host}/dict.json`, true);
    xhrdict.send();

    // 天地图
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
                let mapsuggests_div = document.getElementById("mapsuggests");
                mapsuggests_div.innerHTML = "";
                let suggests = "<ol>";
                for (let i = 0; i < rst_suggests.length; i++) {
                    suggests += `<li>[天地图]&nbsp;<a href="javascript:void(0);" onclick="let element=document.getElementById('rqth');element.value='${rst_suggests[i].address + rst_suggests[i].name}';element.focus();">${rst_suggests[i].name}</a><i>${rst_suggests[i].address}</i></li>`
                }
                suggests += "</ol>"
                mapsuggests_div.innerHTML = suggests;
            }
        }
    });
    document.getElementById("rqth").addEventListener("input", () => {
        local_search.search(document.getElementById("rqth").value, 4);
    });

    // 简写搜索
    // 模式
    document.getElementById("mode").addEventListener("input", () => {
        let ele = document.getElementById("mode");
        ele.value = ele.value.toUpperCase();
        let rst = ALLOWED_MODES.filter((item) => { return item.includes(ele.value); });
        let suggests_div = document.getElementById("suggests");
        suggests_div.innerHTML = "";
        let mapsuggests_div = document.getElementById("mapsuggests");
        mapsuggests_div.innerHTML = "";
        let suggests = "<ol>";
        for (let i in rst) {
            suggests += `<li>[模式]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('mode').value='${rst[i]}';document.getElementById('mode').focus();">${rst[i]}</a></li>`;
        }
        suggests += "</ol>";
        suggests_div.innerHTML = suggests;
    });
    // 呼号 - 模板
    document.getElementById("callsign").addEventListener("input", () => {
        let ele = document.getElementById("callsign");
        ele.value = ele.value.toUpperCase();
        let rst = tmpljson.filter((item) => { return item.callsign.includes(ele.value); });
        let suggests_div = document.getElementById("suggests");
        suggests_div.innerHTML = "";
        let mapsuggests_div = document.getElementById("mapsuggests");
        mapsuggests_div.innerHTML = "";
        let suggests = "<ol>";
        for (let i in rst) {
            suggests += `<li>[模板]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('callsign').value='${rst[i].callsign}';document.getElementById('rrig').value='${rst[i].rig}';document.getElementById('rpwr').value='${rst[i].pwr}';document.getElementById('rant').value='${rst[i].ant}';document.getElementById('rqth').value='${rst[i].qth}';document.getElementById('callsign').focus();">${rst[i].callsign}</a><i>${rst[i].rig}|${rst[i].pwr}|${rst[i].ant}|${rst[i].qth}</i></li>`;
        }
        suggests += "</ol>";
        suggests_div.innerHTML = suggests;
    });
    // 设备 - 字典
    document.getElementById("rrig").addEventListener("input", () => {
        let rst = [];
        let rstkeys = Object.keys(dictjson.rig).filter((item) => { return item.includes(document.getElementById("rrig").value.toLowerCase()); });
        for (let i in rstkeys) {
            rst = rst.concat(dictjson.rig[rstkeys[i]]);
        }
        if (rst) {
            let suggests_div = document.getElementById("suggests");
            suggests_div.innerHTML = "";
            let mapsuggests_div = document.getElementById("mapsuggests");
            mapsuggests_div.innerHTML = "";
            let suggests = "<ol>";
            for (let i in rst) {
                suggests += `<li>[设备]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('rrig').value='${rst[i]}';document.getElementById('rrig').focus();">${rst[i]}</a></li>`
            }
            suggests += "</ol>";
            suggests_div.innerHTML = suggests;
        }
    });
    // 功率 - 字典
    document.getElementById("rpwr").addEventListener("input", () => {
        let rst = [];
        let rstkeys = Object.keys(dictjson.pwr).filter((item) => { return item.includes(document.getElementById("rpwr").value.toLowerCase()); });
        for (let i in rstkeys) {
            rst = rst.concat(dictjson.pwr[rstkeys[i]]);
        }
        if (rst) {
            let suggests_div = document.getElementById("suggests");
            suggests_div.innerHTML = "";
            let mapsuggests_div = document.getElementById("mapsuggests");
            mapsuggests_div.innerHTML = "";
            let suggests = "<ol>";
            for (let i in rst) {
                suggests += `<li>[功率]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('rpwr').value='${rst[i]}';document.getElementById('rpwr').focus();">${rst[i]}</a></li>`
            }
            suggests += "</ol>";
            suggests_div.innerHTML = suggests;
        }
    });
    // 天线 - 字典
    document.getElementById("rant").addEventListener("input", () => {
        let rst = [];
        let rstkeys = Object.keys(dictjson.ant).filter((item) => { return item.includes(document.getElementById("rant").value.toLowerCase()); });
        for (let i in rstkeys) {
            rst = rst.concat(dictjson.ant[rstkeys[i]]);
        }
        if (rst) {
            let suggests_div = document.getElementById("suggests");
            suggests_div.innerHTML = "";
            let mapsuggests_div = document.getElementById("mapsuggests");
            mapsuggests_div.innerHTML = "";
            let suggests = "<ol>";
            for (let i in rst) {
                suggests += `<li>[天线]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('rant').value='${rst[i]}';document.getElementById('rant').focus();">${rst[i]}</a></li>`
            }
            suggests += "</ol>";
            suggests_div.innerHTML = suggests;
        }
    });
    // QTH - 字典
    document.getElementById("rqth").addEventListener("input", () => {
        let rst = [];
        let rstkeys = Object.keys(dictjson.qth).filter((item) => { return item.includes(document.getElementById("rqth").value.toLowerCase()); });
        for (let i in rstkeys) {
            rst = rst.concat(dictjson.qth[rstkeys[i]]);
        }
        if (rst) {
            let suggests_div = document.getElementById("suggests");
            suggests_div.innerHTML = "";
            let mapsuggests_div = document.getElementById("mapsuggests");
            mapsuggests_div.innerHTML = "";
            let suggests = "<ol>";
            for (let i in rst) {
                suggests += `<li>[QTH]&nbsp;<a href="javascript:void(0);" onclick="document.getElementById('rqth').value='${rst[i]}';document.getElementById('rqth').focus();">${rst[i]}</a></li>`
            }
            suggests += "</ol>";
            suggests_div.innerHTML = suggests;
        }
    });

    // 检查复选框状态
    document.getElementById("dt").disabled = document.getElementById("dtauto").checked;
    let lock_names = ["freq", "mode", "trig", "tpwr", "tant", "tqth"];
    for (let i in lock_names) document.getElementById(lock_names[i]).disabled = document.getElementById(lock_names[i] + "lock").checked;

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
        socket.send(JSON.stringify(retjson));
        clear_input();
        document.getElementById("callsign").focus();
    });

    // 询问文件名
    logname = prompt("输入要打开的文件名（不存在则新建）", "newlogbook");

    // ws
    socket = new WebSocket(`ws://${window.location.host}/ws`);
    socket.onopen = () => { socket.send(JSON.stringify({ "type": 1, "message": logname })); };
    socket.onmessage = (ev) => {
        let info = JSON.parse(String(ev.data));
        if (info.type == 2 && info.message == "OK") {
            document.getElementById("logtitle").innerText = "记录 | 服务已连接";
        } else if (info.type == 3) {
            // 更新提交提示
            nextlog = nextlog > info.payload.index ? nextlog : info.payload.index + 1;
            document.getElementById("submit").value = `提交 #${nextlog}`;
            // 记录表格
            let keys = ["callsign", "dt", "freq", "mode", "rst", "rrig", "rpwr", "rant", "rqth", "trig", "tpwr", "tant", "tqth", "rmks"];
            let inner = `<td id="log_td_i${info.payload.index}_index"><a href="javascript:void(0);" onclick="editlog(${info.payload.index})">${info.payload.index}</a></td>`;
            for (let i = 0; i < keys.length; i++) {
                inner += `<td id="log_td_i${info.payload.index}_${keys[i]}">${info.payload[keys[i]]}</td>`;
            }
            if (info.message == "SYNC" || info.message == "ADD") {
                let newtr = document.createElement("tr");
                newtr.id = `log_tr_i${info.payload.index}`;
                newtr.innerHTML = inner;
                document.getElementById("logtable").appendChild(newtr);
            } else if (info.message == "EDIT") {
                document.getElementById(`log_tr_i${info.payload.index}`).innerHTML = inner;
            }
            // 自动滚动
            let logdiv = document.getElementById("logs");
            logdiv.scroll({
                top: logdiv.scrollHeight,
                left: 0,
                behavior: "smooth",
            });
            if (info.message == "ADD" || info.message == "EDIT") {
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
