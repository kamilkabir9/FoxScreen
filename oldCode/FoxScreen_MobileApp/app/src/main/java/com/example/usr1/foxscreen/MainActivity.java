package com.example.usr1.foxscreen;

import android.app.DownloadManager;
import android.content.BroadcastReceiver;
import android.content.Intent;
import android.content.IntentFilter;
import android.graphics.Point;
import android.net.Uri;
import android.os.Environment;
import android.support.v7.app.AppCompatActivity;
import android.content.Context;
import android.os.Bundle;
import android.net.wifi.WifiInfo;
import android.net.wifi.WifiManager;
import android.util.Log;
import android.util.Base64;
import android.view.Display;
import android.view.View;
import android.widget.Button;
import android.widget.Toast;

import com.android.volley.Request;
import com.android.volley.RequestQueue;
import com.android.volley.Response;
import com.android.volley.VolleyError;
import com.android.volley.toolbox.StringRequest;
import com.android.volley.toolbox.Volley;
import com.google.gson.Gson;
//    private void HandShake() {
//
//        final String HSurl = wsURL+"/HandShake";
//
//        try {
//            mConnection.connect(HSurl, new WebSocketHandler() {
//
//                @Override
//                public void onOpen() {
//                    Log.d(TAG, "Status: Connected to " + HSurl);
//                    String fakeJson=DeviceJson();
//                    Log.d(TAG,"sending Msg:"+fakeJson);
//                    mConnection.sendBinaryMessage(fakeJson.getBytes());
//                }
//
//                @Override
//                public void onTextMessage(String payload) {
//                    Log.d(TAG, "Got echo: " + payload);
//
//                }
//
//                @Override
//                public void onClose(int code, String reason) {
//                    Log.d(TAG, "Connection lost.");
//                }
//            });
//        } catch (WebSocketException e) {
//
//            Log.d(TAG, e.toString());
//        }
//    }
import java.io.File;

import de.tavendo.autobahn.WebSocketConnection;

public class MainActivity extends AppCompatActivity {
    RequestQueue queue ;
    String mac64;
    Gson gson=new Gson();
    final String rURL="http://192.168.1.102:8080";
//    final String rURL="http://1ed039a3.ngrok.io";
    private static final String TAG = "FXTag";
    private static final String KnockTag = "FXTKnock";
    boolean KnockResult=false;
//    private final WebSocketConnection mConnection = new WebSocketConnection();
    private boolean Knock(String loc){
        long epoch=System.nanoTime()/1000;
        Log.d(KnockTag,String.format("Knock IN %s @ %s",loc,epoch));
        final String Kurl = rURL+"/Knock";
// Request a string response from the provided URL.
        String urlParam= KurlParam(loc,epoch);
//        Log.d(KnockTag,"url:"+urlParam);
        StringRequest stringRequest = new StringRequest(Request.Method.GET, Kurl+urlParam,
                new Response.Listener<String>() {
                    @Override
                    public void onResponse(String response) {
                        // Display the first 500 characters of the response string.
                        Log.d(TAG,("Response is: "+ response.toString()));
                        RESTrply rply=gson.fromJson(response.toString(),RESTrply.class);
                        KnockResult=rply.Err;
                    }
                }, new Response.ErrorListener() {
            @Override
            public void onErrorResponse(VolleyError error) {
                Log.d(TAG,"That Knock didn't work!");
                Log.d(TAG,error.toString());
                KnockResult=false;
            }
        });
// Add the request to the RequestQueue.
        queue.add(stringRequest);
        return KnockResult;
    }

    private boolean HandShake(){
//        TODO check if wifi is on DEvise
// Instantiate the RequestQueue.
        final String HSurl = rURL+"/HandShake";
// Request a string response from the provided URL.
        String urlParam= DeviceHSurlParam();
        StringRequest stringRequest = new StringRequest(Request.Method.GET, HSurl+urlParam,
                new Response.Listener<String>() {
                    @Override
                    public void onResponse(String response) {
                        // Display the first 500 characters of the response string.
                        Log.d(TAG,("Response is: "+ response.toString()));
                        RESTrply rply=gson.fromJson(response.toString(),RESTrply.class);
                        KnockResult=rply.Err;
                    }
                }, new Response.ErrorListener() {
            @Override
            public void onErrorResponse(VolleyError error) {
                Log.d(TAG,"That HandShake didn't work!");
                Log.d(TAG,error.toString());
//                Toast.makeText(this, "That HandShake didn't work!",Toast.LENGTH_LONG).show();
            }
        });
// Add the request to the RequestQueue.
        queue.add(stringRequest);
        return KnockResult;
    }
    public void DownlaodMP4(){
        long downloadReference;
        Uri video_uri = Uri.parse("http://192.168.1.102:8080/DownloadMp4?FileID=2e2aac39-e07e-11e6-bcb6-2c27d7cd4a43.mp4");
        // Create request for android download manager
        DownloadManager downloadManager = (DownloadManager)getSystemService(DOWNLOAD_SERVICE);
        DownloadManager.Request request = new DownloadManager.Request(video_uri);

        //Setting title of request
        request.setTitle("FoxScreen.mp4");

        //Setting description of request
        request.setDescription("FoxSCreen mp4 file");

        //Set the local destination for the downloaded file to a path
        //within the application's external files directory
        request.setDestinationInExternalPublicDir("/FoxScreen","video.mp4");
        registerReceiver(onComplete, new IntentFilter(DownloadManager.ACTION_DOWNLOAD_COMPLETE));
        //Enqueue download and save into referenceId
        downloadReference = downloadManager.enqueue(request);
        Log.d(TAG, String.valueOf(downloadReference)+" downloadReference Number");

//        Button DownloadStatus = (Button) findViewById(R.id.DownloadStatus);
//        DownloadStatus.setEnabled(true);
//        Button CancelDownload = (Button) findViewById(R.id.CancelDownload);
//        CancelDownload.setEnabled(true);


    }
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
//        DownlaodMP4();
        queue= Volley.newRequestQueue(this);
        Log.d(TAG, DeviceHSurlParam());
        Log.d(TAG,"HandShake worked :"+HandShake());
        final Button button_North = (Button) findViewById(R.id.North);
        final Button button_West = (Button) findViewById(R.id.West);
        final Button button_East = (Button) findViewById(R.id.East);
        final Button button_South = (Button) findViewById(R.id.South);

        button_North.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                String loc="North";
                Log.d(TAG,"Knock "+loc);
                Knock(loc);
            }
        });

        button_West.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                String loc="West";
                Log.d(TAG,"Knock "+loc);
                Knock(loc);
            }
        });

        button_East.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                String loc="East";
                Log.d(TAG,"Knock "+loc);
                Knock(loc);
            }
        });

        button_South.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                String loc="South";
                Log.d(TAG,"Knock "+loc);
                Knock(loc);
            }
        });

    }
    public String DeviceHSurlParam(){

        WifiManager manager = (WifiManager) getApplicationContext().getSystemService(Context.WIFI_SERVICE);
        WifiInfo info = manager.getConnectionInfo();
        String macAddress = info.getMacAddress();
        mac64=Base64.encodeToString(macAddress.getBytes(),Base64.NO_WRAP);
        Log.i(TAG, "mac = " + macAddress);

        //display Name
        Display display = getWindowManager().getDefaultDisplay();
        String displayName = display.getName();  // minSdkVersion=17+
//        Log.i(TAG, "displayName  = " + displayName);

        // display size in pixels
        Point size = new Point();
        display.getSize(size);
        int width = size.x;
        int height = size.y;
//        Log.i(TAG, "width        = " + width);
//        Log.i(TAG, "height       = " + height);

        // pixels, dpi
//        DisplayMetrics metrics = new DisplayMetrics();
//        getWindowManager().getDefaultDisplay().getMetrics(metrics);
//        int heightPixels = metrics.heightPixels;
//        int widthPixels = metrics.widthPixels;
//        int densityDpi = metrics.densityDpi;
//        float xdpi = metrics.xdpi;
//        float ydpi = metrics.ydpi;
//        Log.i(TAG, "widthPixels  = " + widthPixels);
//        Log.i(TAG, "heightPixels = " + heightPixels);
//        Log.i(TAG, "densityDpi   = " + densityDpi);
//        Log.i(TAG, "xdpi         = " + xdpi);
//        Log.i(TAG, "ydpi         = " + ydpi);
//

        // orientation (either ORIENTATION_LANDSCAPE, ORIENTATION_PORTRAIT)
        int orientation = getResources().getConfiguration().orientation;
//        Log.i(TAG, "orientation  = " + orientation);

        String urlParam=String.format("?Mac=%s&Width=%s&Height=%s&Orientation=%s", mac64,width,height,orientation);
        return urlParam;

    }
    public String KurlParam(String loc,long epoch){
        String urlParam=String.format("?Mac=%s&Loc=%s&Epoch=%s", mac64,loc,epoch);
        return urlParam;
    }

    public class RESTrply{
        boolean Err;
        String Msg;
    }
    BroadcastReceiver onComplete=new BroadcastReceiver() {
        public void onReceive(Context ctxt, Intent intent) {
            Log.d(TAG,"Download complteeeeeeeeeeeeeeee");
            Intent intent1 = new Intent(MainActivity.this,VideoDownload.class);
            startActivity(intent1);
        }
    };

}
