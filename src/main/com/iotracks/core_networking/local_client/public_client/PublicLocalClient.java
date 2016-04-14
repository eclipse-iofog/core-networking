package main.com.iotracks.core_networking.local_client.public_client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.bytes.ByteArrayDecoder;
import io.netty.handler.codec.bytes.ByteArrayEncoder;
import main.com.iotracks.core_networking.local_client.LocalClient;
import main.com.iotracks.core_networking.main.CoreNetworking;
import main.com.iotracks.core_networking.utils.Constants.SocketConnectionStatus;

/**
 * communications in "public" mode
 *
 * Created by saeid on 4/12/16.
 */
public class PublicLocalClient implements LocalClient {
    private final Channel comSatChannel;
    private String localHost;
    private int localPort;
    private Channel ch;
    private Object connectionLock = new Object();
    private SocketConnectionStatus connectionStatus = SocketConnectionStatus.NOT_CONNECTED;

    /**
     * connects to local container
     *
     */
    private Runnable run = () -> {
        EventLoopGroup group = new NioEventLoopGroup(1);
        try {
            Bootstrap b = new Bootstrap();
            b.group(group)
                    .channel(NioSocketChannel.class)
                    .handler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) throws Exception {
                            ch.pipeline().addLast(new ByteArrayDecoder());
                            ch.pipeline().addLast(new ByteArrayEncoder());
                            ch.pipeline().addLast(new PublicLocalClientHandler(comSatChannel));
                        }
                    });

            synchronized (connectionLock) {
                try {
                    ch = b.connect(localHost, localPort).sync().channel();
                    connectionStatus = SocketConnectionStatus.CONNECTED;
                    connectionLock.notifyAll();
                } catch (Exception e) {
                    connectionLock.notifyAll();
                    throw e;
                }
            }
            ch.closeFuture().sync();
            connectionStatus = SocketConnectionStatus.NOT_CONNECTED;
            connectionLock.notifyAll();
        } catch (Exception e) {
            connectionStatus = SocketConnectionStatus.FAILED;
        } finally {
            group.shutdownGracefully();
        }
    };

    public PublicLocalClient(Channel comSatChannel) {
        this.localHost = CoreNetworking.config.getLocalHost();
        this.localPort = CoreNetworking.config.getLocalPort();
        this.comSatChannel = comSatChannel;
    }

    /**
     * pipe received message from ComSat server to local container
     *
     * @param message
     */
    @Override
    public void handleMessage(byte[] message) {
        try {
            ch.writeAndFlush(message).sync();
        } catch (Exception e) {
        }
    }

    @Override
    public boolean isConnected() {
        return connectionStatus.equals(SocketConnectionStatus.CONNECTED);
    }

    @Override
    public boolean connect(long timeout) {
        synchronized (connectionLock) {
            try {
                new Thread(run).start();
                connectionLock.wait(timeout);
            } catch (Exception e) {
            }
        }
        return isConnected();
    }
}
