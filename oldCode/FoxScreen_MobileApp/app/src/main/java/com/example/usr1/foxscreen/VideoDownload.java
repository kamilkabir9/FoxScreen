package com.example.usr1.foxscreen;

import android.content.Intent;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;

public class VideoDownload extends AppCompatActivity {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_video_download);
        Intent i =new Intent(VideoDownload.this,VideoDisplay.class);
        startActivity(i);
    }
}
