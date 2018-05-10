package com.example.canvasandroidexample;

import android.app.Activity;
import android.opengl.*;
import android.os.Bundle;
import android.view.MotionEvent;
import android.view.View;

import javax.microedition.khronos.egl.EGLConfig;
import javax.microedition.khronos.opengles.GL10;

import canvasandroidexample.Canvasandroidexample;

public class MainActivity extends Activity implements GLSurfaceView.Renderer {

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        GLSurfaceView view = new GLSurfaceView(this);
        view.setEGLContextClientVersion(2);
        view.setEGLConfigChooser(8, 8, 8, 8, 0, 8);
        view.setRenderer(this);
        view.setOnTouchListener(new View.OnTouchListener() {
            @Override
            public boolean onTouch(View v, MotionEvent event) {
                int x = Math.round(event.getX());
                int y = Math.round(event.getY());
                if (event.getAction() == MotionEvent.ACTION_DOWN) {
                    Canvasandroidexample.touchEvent("down", x, y);
                } else if (event.getAction() == MotionEvent.ACTION_UP) {
                    Canvasandroidexample.touchEvent("up", x, y);
                } else if (event.getAction() == MotionEvent.ACTION_MOVE) {
                    Canvasandroidexample.touchEvent("move", x, y);
                }
                return true;
            }
        });
        setContentView(view);
    }

    @Override
    protected void onResume() {
        super.onResume();
    }

    @Override
    public void onSurfaceCreated(GL10 gl, EGLConfig config) {
        Canvasandroidexample.onSurfaceCreated();
    }

    @Override
    public void onSurfaceChanged(GL10 gl, int width, int height) {
        Canvasandroidexample.onSurfaceChanged(width, height);
    }

    @Override
    public void onDrawFrame(GL10 gl) {
        Canvasandroidexample.onDrawFrame();
    }
}
