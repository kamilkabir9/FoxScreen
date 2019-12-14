package com.example.usr1.foxscreen;

import android.net.Uri;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.widget.VideoView;

public class VideoDisplay extends AppCompatActivity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_video_display);
        VideoView v=(VideoView) findViewById(R.id.videoView);
//        TODO video name keeps change
        v.setVideoPath("/storage/emulated/0/FoxScreen/v.mp4");
        v.start();
    }
}
