var T = window.T;
var dt_interval;
var map;
var local_search;
var geocoder;
var markers;
var socket;

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
                let suggests_div = document.getElementById("suggests");
                suggests_div.innerHTML = "";
                let suggests = "<ol>";
                for (let i = 0; i < rst_suggests.length; i++) {
                    suggests += `<li><a href="javascript:void(0);" onclick="let element=document.getElementById('rqth');element.value='${rst_suggests[i].name}';element.focus();">${rst_suggests[i].name}</a><i>${rst_suggests[i].address}</i></li>`
                }
                suggests += "</ol>"
                suggests_div.innerHTML = suggests;
            }
        }
    });
    document.getElementById("mapsearchinput").addEventListener("keyup", () => {
        local_search.search(document.getElementById("mapsearchinput").value, 4);
    });

    // 检查复选框状态
    document.getElementById("dt").disabled = document.getElementById("dtauto").checked;
    let lock_names = ["band", "mode", "trig", "tpwr", "tant", "tqth"];
    for (let i in lock_names) document.getElementById(lock_names[i]).disabled = document.getElementById(lock_names[i] + "lock").checked;

    // 清空输入框
    clear_input();

    // ws
    socket = new WebSocket(`ws://${window.location.host}/ws`);
    socket.onopen = () => { socket.send("QSL?"); };
    socket.onmessage = (ev) => {
        let dat = String(ev.data);
        if (dat.startsWith("QSL.")) {
            alert("服务已连接");
            document.getElementById("submit").value = `提交 #${parseInt(dat.substring(4)) + 1}`;
        } else if (dat.startsWith("ADDLOG>")) {
            let info = dat.substring(7);
            // 解析csv (仅限此程序生成的信息)
            let info_arr_orig = info.split(",");
            let info_arr = [""];
            for (let i in info_arr_orig) {
                if (info_arr_orig[i].startsWith('"')) info_arr.push(info_arr_orig[i]);
                else info_arr[info_arr.length - 1] += "," + info_arr_orig[i];
            }
            if (info_arr[0] == "") info_arr.shift();
            if (info_arr.length != 15) alert("CSV信息解析失败"); // 后端补充位次tag
            // 去除首尾引号并解除转义
            for (let i in info_arr) {
                info_arr[i] = info_arr[i].substring(1, info_arr[i].length - 1);
                info_arr[i] = info_arr[i].replace('""', '"');
            }
            document.getElementById("submit").value = `提交 #${parseInt(info_arr[0]) + 1}`;
            /*
                0 - 位次
                1 - 呼号
                2 - 时间
                9 - QTH
            */
            // 记录表格
            let newtr = document.createElement("tr");
            for (let i = 0; i < info_arr.length; i++) {
                let newtd = document.createElement("td");
                newtd.textContent = info_arr[i];
                newtr.appendChild(newtd);
            }
            document.getElementById("logtable").appendChild(newtr);
            // 地理编码
            geocoder.getPoint(info_arr[9], (rst) => {
                if (rst.getStatus() == 0) {
                    map.panTo(rst.getLocationPoint(), 16);
                    let marker = new T.Marker(rst.getLocationPoint());
                    map.addOverLay(marker);
                    let marker_info = new T.InfoWindow(`${info_arr[0]}. ${info_arr[1]}<br>${info_arr[2]}<br>${info_arr[9]}`);
                    marker.addEventListener("click", () => { marker.openInfoWindow(marker_info); });
                    markers.addMarker(marker);
                }
            });
        }
    };
    socket.onerror = () => { alert("服务连接失败"); };

    // 阻止表单提交
    document.getElementById("infoform").addEventListener("submit", (ev) => {
        ev.preventDefault();
        let get_str_and_decorate = (id) => { return `"${document.getElementById(id).value.replace('"', '""')}"`; };
        socket.send(`${get_str_and_decorate("callsign")},${get_str_and_decorate("dt")},${get_str_and_decorate("band")},${get_str_and_decorate("mode")},${get_str_and_decorate("rst")},${get_str_and_decorate("rrig")},${get_str_and_decorate("rpwr")},${get_str_and_decorate("rant")},${get_str_and_decorate("rqth")},${get_str_and_decorate("trig")},${get_str_and_decorate("tpwr")},${get_str_and_decorate("tant")},${get_str_and_decorate("tqth")},${get_str_and_decorate("rmks")}`);
        clear_input();
        document.getElementById("callsign").focus();
    });
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
