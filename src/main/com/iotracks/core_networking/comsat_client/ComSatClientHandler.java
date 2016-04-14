package main.com.iotracks.core_networking.comsat_client;

import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import main.com.iotracks.core_networking.local_client.LocalClient;
import main.com.iotracks.core_networking.local_client.LocalClientBuilder;

import java.util.logging.Logger;

/**
 * Created by saeid on 4/8/16.
 */
public class ComSatClientHandler extends ChannelInboundHandlerAdapter {

    private final Logger log = Logger.getLogger(ComSatClientHandler.class.getName());
    private final ComSatClient client;
    private LocalClient localClient = null;

    public ComSatClientHandler(ComSatClient client) {
        this.client = client;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        if (localClient == null)
            localClient = LocalClientBuilder.build(client.getChannel());
    }

    @Override
    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
        byte[] contentBytes = ((byte[]) msg);
        String contentString = "";
        if (contentBytes.length < 11) {
            contentString = new String(contentBytes);
            if (contentString.equals("BEAT") || contentString.equals("AUTHORIZED") || contentString.equals("BEATBEAT")) {
                client.seen();
                return;
            }
        }

        if (!localClient.isConnected()) {
            if (!localClient.connect(2000)) {
                log.warning("unable to connect to local client");
                return;
            }
        }

        localClient.sendMessage(contentBytes);

    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        log.warning(String.format("#%d : %s", client.getConnectionId(), cause.getMessage()));
        ctx.close();
    }

}
