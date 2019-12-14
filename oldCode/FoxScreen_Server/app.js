/**
 * Created by usr1 on 1/18/17.
 */
function GetDeviceList()
{
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.open( "GET", "/WebView/DeviceList", false ); // false for synchronous request
    xmlHttp.send();
    var res=JSON.parse(xmlHttp.responseText);
    console.log(res);
    if (res.Type="ok"){
        alert("got devices");
    }
}

function GetKnockList()
{
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.open( "GET", "/WebView/KnocksList", false ); // false for synchronous request
    xmlHttp.send();
    var res=JSON.parse(xmlHttp.responseText);
    console.log(res);
    if (res.Type="ok"){
        alert("got knocks");
    }

}