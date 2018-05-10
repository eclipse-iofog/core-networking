package org.eclipse.iofog.core.networking.local.client.public_client;

import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.bytes.ByteArrayDecoder;
import io.netty.handler.codec.bytes.ByteArrayEncoder;
import org.eclipse.iofog.core.networking.local.client.LocalClient;
import org.eclipse.iofog.core.networking.main.CoreNetworking;

import java.util.logging.Logger;

/**
 * communications in "public" mode
 * <p>
 * Created by saeid on 4/12/16.
 */
public class PublicLocalClient implements LocalClient {
    private final Logger log = Logger.getLogger(PublicLocalClient.class.getName());
    private final Channel comSatChannel;
    private String localHost;
    private int localPort;
    private Channel ch;
    /**
     * connects to local container
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

            log.info("connecting to local client");
            ch = b.connect(localHost, localPort).sync().channel();
            log.info("connected to local client");
            ch.closeFuture().sync();
            log.warning("connection to local client closed");
        } catch (Exception e) {
            ch = null;
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
            log.info("sending bytes to local client");
            ch.writeAndFlush(message).sync();
        } catch (Exception e) {
        }
    }

    @Override
    public boolean isConnected() {
        return (ch != null && ch.isActive());
    }

    @Override
    public void closeConnection() {
        if (ch != null)
            try {
                ch.disconnect();
                ch.close();
            } catch (Exception e) {
            }
    }

    @Override
    public boolean connect(long timeout) {
        new Thread(run).start();
        while (!isConnected()) {
            try {
                Thread.sleep(20);
            } catch (Exception e) {
            }
        }
        return isConnected();
    }
}
