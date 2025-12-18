var T = window.T;
var dt_interval;
var map;
var local_search;
var geocoder;
var markers;
var socket;
var logname;

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
