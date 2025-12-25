var T = window.T;
var dt_interval;
var map;
var local_search;
var geocoder;
var markers;
var socket;
var logname;
var tmpljson, dictjson;

// 时间更新函数
function update_dt() {
    const dt = document.getElementById("dt");
    const now = new Date();
    dt.value = now.toISOString().substring(0, 16);
}

// 输入框清理函数
function clear_input() {
    document.querySelectorAll("input").forEach((ele) => { if (ele.type == "text" && !ele.disabled) ele.value = ""; });
    document.getElementById("rst").value = 59; // by default
}

function onload() {
    // 时间同步更新
    update_dt();

    // 天地图
    map = new T.Map("mapdiv");
    map.centerAndZoom(new T.LngLat(116.40769, 39.89945), 12);
    geocoder = new T.Geocoder();
    markers = new T.MarkerClusterer(map, { markers: [] });

    // 获取记忆数据
    let xhrtmpl = new XMLHttpRequest();
    xhrtmpl.onreadystatechange = () => {
        if (xhrtmpl.readyState == 4 && xhrtmpl.status == 200) {
            tmpljson = JSON.parse(xhrtmpl.responseText);
        }
    };
    xhrtmpl.open("GET", `http://${window.location.host}/tmpl.json`, true);
    xhrtmpl.send();
    let xhrdict = new XMLHttpRequest();
    xhrdict.onreadystatechange = () => {
        if (xhrdict.readyState == 4 && xhrdict.status == 200) {
            dictjson = JSON.parse(xhrdict.responseText);
        }
    };
    xhrdict.open("GET", `http://${window.location.host}/dict.json`, true);
    xhrdict.send();

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
    document.getElementById("callsign").addEventListener("input", () => {
        let ele = document.getElementById("callsign");
        ele.value = ele.value.toUpperCase()
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
    document.getElementById("rrig").addEventListener("input", () => {
        let rst = dictjson.rig[document.getElementById("rrig").value.toLowerCase()];
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
    document.getElementById("rpwr").addEventListener("input", () => {
        let rst = dictjson.pwr[document.getElementById("rpwr").value.toLowerCase()];
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
    document.getElementById("rant").addEventListener("input", () => {
        let rst = dictjson.ant[document.getElementById("rant").value.toLowerCase()];
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
    document.getElementById("rqth").addEventListener("input", () => {
        let rst = dictjson.qth[document.getElementById("rqth").value.toLowerCase()];
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
    let lock_names = ["band", "mode", "trig", "tpwr", "tant", "tqth"];
    for (let i in lock_names) document.getElementById(lock_names[i]).disabled = document.getElementById(lock_names[i] + "lock").checked;

    // 清空输入框
    clear_input();

    // 表单提交行为
    document.getElementById("infoform").addEventListener("submit", (ev) => {
        ev.preventDefault();
        let info_json = {
            "callsign": document.getElementById("callsign").value,
            "dt": document.getElementById("dt").value,
            "band": document.getElementById("band").value,
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
        socket.send(JSON.stringify(info_json));
        clear_input();
        document.getElementById("callsign").focus();
    });

    // 询问文件名
    logname = prompt("输入要打开的文件名（不存在则新建）", "newlogbook");

    // ws
    socket = new WebSocket(`ws://${window.location.host}/ws`);
    socket.onopen = () => { socket.send("QSL?" + logname); };
    socket.onmessage = (ev) => {
        let dat = String(ev.data);
        if (dat == "QSL.") {
            alert("服务已连接");
        } else {
            let info = JSON.parse(dat);
            document.getElementById("submit").value = `提交 #${info.index + 1}`;
            // 记录表格
            let keys = ["index", "callsign", "dt", "band", "mode", "rst", "rrig", "rpwr", "rant", "rqth", "trig", "tpwr", "tant", "tqth", "rmks"];
            let newtr = document.createElement("tr");
            for (let i = 0; i < keys.length; i++) {
                let newtd = document.createElement("td");
                newtd.textContent = info[keys[i]];
                newtr.appendChild(newtd);
            }
            document.getElementById("logtable").appendChild(newtr);
            // 自动滚动
            let logdiv = document.getElementById("logs");
            logdiv.scroll({
                top: logdiv.scrollHeight,
                left: 0,
                behavior: "smooth",
            });
            // 地理编码
            geocoder.getPoint(info.rqth, (rst) => {
                if (rst.getStatus() == 0) {
                    map.panTo(rst.getLocationPoint(), 16);
                    let marker = new T.Marker(rst.getLocationPoint());
                    map.addOverLay(marker);
                    let marker_info = new T.InfoWindow(`${info.index}. ${info.callsign}<br>${info.dt}<br>${info.rqth}`);
                    marker.addEventListener("click", () => { marker.openInfoWindow(marker_info); });
                    markers.addMarker(marker);
                }
            });
        }
    };
    socket.onerror = () => { alert("服务连接失败"); };
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
